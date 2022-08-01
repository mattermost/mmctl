// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-server/v6/audit"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

func (api *API) InitCloud() {
	// GET /api/v4/cloud/products
	api.BaseRoutes.Cloud.Handle("/products", api.APISessionRequired(getCloudProducts)).Methods("GET")
	// GET /api/v4/cloud/limits
	api.BaseRoutes.Cloud.Handle("/limits", api.APISessionRequired(getCloudLimits)).Methods("GET")

	// POST /api/v4/cloud/payment
	// POST /api/v4/cloud/payment/confirm
	api.BaseRoutes.Cloud.Handle("/payment", api.APISessionRequired(createCustomerPayment)).Methods("POST")
	api.BaseRoutes.Cloud.Handle("/payment/confirm", api.APISessionRequired(confirmCustomerPayment)).Methods("POST")

	// GET /api/v4/cloud/customer
	// PUT /api/v4/cloud/customer
	// PUT /api/v4/cloud/customer/address
	api.BaseRoutes.Cloud.Handle("/customer", api.APISessionRequired(getCloudCustomer)).Methods("GET")
	api.BaseRoutes.Cloud.Handle("/customer", api.APISessionRequired(updateCloudCustomer)).Methods("PUT")
	api.BaseRoutes.Cloud.Handle("/customer/address", api.APISessionRequired(updateCloudCustomerAddress)).Methods("PUT")

	// GET /api/v4/cloud/subscription
	api.BaseRoutes.Cloud.Handle("/subscription", api.APISessionRequired(getSubscription)).Methods("GET")
	api.BaseRoutes.Cloud.Handle("/subscription/invoices", api.APISessionRequired(getInvoicesForSubscription)).Methods("GET")
	api.BaseRoutes.Cloud.Handle("/subscription/invoices/{invoice_id:in_[A-Za-z0-9]+}/pdf", api.APISessionRequired(getSubscriptionInvoicePDF)).Methods("GET")
	api.BaseRoutes.Cloud.Handle("/subscription", api.APISessionRequired(changeSubscription)).Methods("PUT")

	// GET /api/v4/cloud/request-trial
	api.BaseRoutes.Cloud.Handle("/request-trial", api.APISessionRequired(requestCloudTrial)).Methods("PUT")

	// GET /api/v4/cloud/validate-business-email
	api.BaseRoutes.Cloud.Handle("/validate-business-email", api.APISessionRequired(validateBusinessEmail)).Methods("POST")

	// POST /api/v4/cloud/webhook
	api.BaseRoutes.Cloud.Handle("/webhook", api.CloudAPIKeyRequired(handleCWSWebhook)).Methods("POST")
}

func getSubscription(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.getSubscription", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleReadBilling) {
		c.SetPermissionError(model.PermissionSysconsoleReadBilling)
		return
	}

	subscription, err := c.App.Cloud().GetSubscription(c.AppContext.Session().UserId)
	if err != nil {
		c.Err = model.NewAppError("Api4.getSubscription", "api.cloud.request_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(subscription)
	if err != nil {
		c.Err = model.NewAppError("Api4.getSubscription", "api.cloud.request_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(json)
}

func changeSubscription(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.changeSubscription", "api.cloud.license_error", nil, "", http.StatusInternalServerError)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleWriteBilling) {
		c.SetPermissionError(model.PermissionSysconsoleWriteBilling)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Err = model.NewAppError("Api4.changeSubscription", "api.cloud.app_error", nil, err.Error(), http.StatusBadRequest)
		return
	}

	var subscriptionChange *model.SubscriptionChange
	if err = json.Unmarshal(bodyBytes, &subscriptionChange); err != nil {
		c.Err = model.NewAppError("Api4.changeSubscription", "api.cloud.app_error", nil, err.Error(), http.StatusBadRequest)
		return
	}

	currentSubscription, appErr := c.App.Cloud().GetSubscription(c.AppContext.Session().UserId)
	if appErr != nil {
		c.Err = model.NewAppError("Api4.changeSubscription", "api.cloud.app_error", nil, appErr.Error(), http.StatusInternalServerError)
		return
	}

	changedSub, err := c.App.Cloud().ChangeSubscription(c.AppContext.Session().UserId, currentSubscription.ID, subscriptionChange)
	if err != nil {
		c.Err = model.NewAppError("Api4.changeSubscription", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(changedSub)
	if err != nil {
		c.Err = model.NewAppError("Api4.changeSubscription", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log failures for purchase confirmation email, but don't show an error to the user so as not to confuse them
	// At this point, the upgrade is complete.
	if nErr := c.App.SendUpgradeConfirmationEmail(); nErr != nil {
		c.Logger.Error("Error sending purchase confirmation email")
	}

	w.Write(json)
}

func requestCloudTrial(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.requestCloudTrial", "api.cloud.license_error", nil, "", http.StatusForbidden)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleWriteBilling) {
		c.SetPermissionError(model.PermissionSysconsoleWriteBilling)
		return
	}

	// check if the email needs to be set
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Err = model.NewAppError("Api4.requestCloudTrial", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}
	// this value will not be empty when both emails (user admin and CWS customer) are not business email and
	// we need to request a new email from the user via the request business email modal
	var startTrialRequest *model.StartCloudTrialRequest
	if err = json.Unmarshal(bodyBytes, &startTrialRequest); err != nil {
		c.Err = model.NewAppError("Api4.requestCloudTrial", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	changedSub, err := c.App.Cloud().RequestCloudTrial(c.AppContext.Session().UserId, startTrialRequest.SubscriptionID, startTrialRequest.Email)
	if err != nil {
		c.Err = model.NewAppError("Api4.requestCloudTrial", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(changedSub)
	if err != nil {
		c.Err = model.NewAppError("Api4.requestCloudTrial", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	defer c.App.Srv().Cloud.InvalidateCaches()

	w.Write(json)
}

func validateBusinessEmail(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.validateBusinessEmail", "api.cloud.license_error", nil, "", http.StatusForbidden)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleWriteBilling) {
		c.SetPermissionError(model.PermissionSysconsoleWriteBilling)
		return
	}

	user, userErr := c.App.GetUser(c.AppContext.Session().UserId)
	if userErr != nil {
		c.Err = model.NewAppError("Api4.validateBusinessEmail", "api.cloud.request_error", nil, "", http.StatusInternalServerError)
		return
	}

	// if an email was sent as a body param, validate it and return wether is valid or not
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Err = model.NewAppError("Api4.requestCloudTrial", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	var emailToValidate *model.ValidateBusinessEmailRequest
	if err := json.Unmarshal(bodyBytes, &emailToValidate); err != nil {
		c.Err = model.NewAppError("Api4.requestCloudTrial", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	if emailToValidate.Email != "" {
		errValidatingEmail := c.App.Cloud().ValidateBusinessEmail(user.Id, emailToValidate.Email)
		if errValidatingEmail != nil {
			c.Err = model.NewAppError("Api4.valiateBusinessEmail", "api.cloud.request_error", nil, errValidatingEmail.Error(), http.StatusInternalServerError)
			return
		}
		ReturnStatusOK(w)
		return
	}

	// If no email was sent as body param, then validate current userAdmin email
	errValidatingAdminEmail := c.App.Cloud().ValidateBusinessEmail(user.Id, user.Email)

	// if the current admin email is not a valid email
	if errValidatingAdminEmail != nil {

		// get the cloud customer email
		cloudCustomer, err := c.App.Cloud().GetCloudCustomer(user.Id)
		if err != nil {
			c.Err = model.NewAppError("Api4.valiateBusinessEmail", "api.cloud.request_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		// and validate that one
		errValidatingSystemEmail := c.App.Cloud().ValidateBusinessEmail(user.Id, cloudCustomer.Email)
		if errValidatingSystemEmail != nil {
			c.Err = model.NewAppError("Api4.valiateBusinessEmail", "api.cloud.request_error", nil, errValidatingSystemEmail.Error(), http.StatusInternalServerError)
			return
		}
	}

	// if any of the emails is valid, return ok
	ReturnStatusOK(w)
}

func getCloudProducts(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.getCloudProducts", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleReadBilling) {
		c.SetPermissionError(model.PermissionSysconsoleReadBilling)
		return
	}

	includeLegacyProducts := r.URL.Query().Get("include_legacy") == "true"

	products, err := c.App.Cloud().GetCloudProducts(c.AppContext.Session().UserId, includeLegacyProducts)

	if err != nil {
		c.Err = model.NewAppError("Api4.getCloudProducts", "api.cloud.request_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(products)
	if err != nil {
		c.Err = model.NewAppError("Api4.getCloudProducts", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(json)
}

func getCloudLimits(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.getCloudLimits", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	limits, err := c.App.Cloud().GetCloudLimits(c.AppContext.Session().UserId)
	if err != nil {
		c.Err = model.NewAppError("Api4.getCloudLimits", "api.cloud.request_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(limits)
	if err != nil {
		c.Err = model.NewAppError("Api4.getCloudLimits", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(json)
}

func getCloudCustomer(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.getCloudCustomer", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleReadBilling) {
		c.SetPermissionError(model.PermissionSysconsoleReadBilling)
		return
	}

	customer, err := c.App.Cloud().GetCloudCustomer(c.AppContext.Session().UserId)
	if err != nil {
		c.Err = model.NewAppError("Api4.getCloudCustomer", "api.cloud.request_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(customer)
	if err != nil {
		c.Err = model.NewAppError("Api4.getCloudCustomer", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(json)
}

func updateCloudCustomer(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.updateCloudCustomer", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleWriteBilling) {
		c.SetPermissionError(model.PermissionSysconsoleWriteBilling)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Err = model.NewAppError("Api4.updateCloudCustomer", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	var customerInfo *model.CloudCustomerInfo
	if err = json.Unmarshal(bodyBytes, &customerInfo); err != nil {
		c.Err = model.NewAppError("Api4.updateCloudCustomer", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	customer, appErr := c.App.Cloud().UpdateCloudCustomer(c.AppContext.Session().UserId, customerInfo)
	if appErr != nil {
		c.Err = model.NewAppError("Api4.updateCloudCustomer", "api.cloud.request_error", nil, appErr.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(customer)
	if err != nil {
		c.Err = model.NewAppError("Api4.updateCloudCustomer", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(json)
}

func updateCloudCustomerAddress(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.updateCloudCustomerAddress", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleWriteBilling) {
		c.SetPermissionError(model.PermissionSysconsoleWriteBilling)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Err = model.NewAppError("Api4.updateCloudCustomerAddress", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	var address *model.Address
	if err = json.Unmarshal(bodyBytes, &address); err != nil {
		c.Err = model.NewAppError("Api4.updateCloudCustomerAddress", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	customer, appErr := c.App.Cloud().UpdateCloudCustomerAddress(c.AppContext.Session().UserId, address)
	if appErr != nil {
		c.Err = model.NewAppError("Api4.updateCloudCustomerAddress", "api.cloud.request_error", nil, appErr.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(customer)
	if err != nil {
		c.Err = model.NewAppError("Api4.updateCloudCustomerAddress", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(json)
}

func createCustomerPayment(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.createCustomerPayment", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleWriteBilling) {
		c.SetPermissionError(model.PermissionSysconsoleWriteBilling)
		return
	}

	auditRec := c.MakeAuditRecord("createCustomerPayment", audit.Fail)
	defer c.LogAuditRec(auditRec)

	intent, err := c.App.Cloud().CreateCustomerPayment(c.AppContext.Session().UserId)
	if err != nil {
		c.Err = model.NewAppError("Api4.createCustomerPayment", "api.cloud.request_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(intent)
	if err != nil {
		c.Err = model.NewAppError("Api4.createCustomerPayment", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	auditRec.Success()

	w.Write(json)
}

func confirmCustomerPayment(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.confirmCustomerPayment", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleWriteBilling) {
		c.SetPermissionError(model.PermissionSysconsoleWriteBilling)
		return
	}

	auditRec := c.MakeAuditRecord("confirmCustomerPayment", audit.Fail)
	defer c.LogAuditRec(auditRec)

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Err = model.NewAppError("Api4.confirmCustomerPayment", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	var confirmRequest *model.ConfirmPaymentMethodRequest
	if err = json.Unmarshal(bodyBytes, &confirmRequest); err != nil {
		c.Err = model.NewAppError("Api4.confirmCustomerPayment", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.App.Cloud().ConfirmCustomerPayment(c.AppContext.Session().UserId, confirmRequest)
	if err != nil {
		c.Err = model.NewAppError("Api4.createCustomerPayment", "api.cloud.request_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	auditRec.Success()

	ReturnStatusOK(w)
}

func getInvoicesForSubscription(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.getInvoicesForSubscription", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleReadBilling) {
		c.SetPermissionError(model.PermissionSysconsoleReadBilling)
		return
	}

	invoices, appErr := c.App.Cloud().GetInvoicesForSubscription(c.AppContext.Session().UserId)
	if appErr != nil {
		c.Err = model.NewAppError("Api4.getInvoicesForSubscription", "api.cloud.request_error", nil, appErr.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(invoices)
	if err != nil {
		c.Err = model.NewAppError("Api4.getInvoicesForSubscription", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(json)
}

func getSubscriptionInvoicePDF(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.getSubscriptionInvoicePDF", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	c.RequireInvoiceId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionSysconsoleReadBilling) {
		c.SetPermissionError(model.PermissionSysconsoleReadBilling)
		return
	}

	pdfData, filename, appErr := c.App.Cloud().GetInvoicePDF(c.AppContext.Session().UserId, c.Params.InvoiceId)
	if appErr != nil {
		c.Err = model.NewAppError("Api4.getSubscriptionInvoicePDF", "api.cloud.request_error", nil, appErr.Error(), http.StatusInternalServerError)
		return
	}

	writeFileResponse(
		filename,
		"application/pdf",
		int64(binary.Size(pdfData)),
		time.Now(),
		*c.App.Config().ServiceSettings.WebserverMode,
		bytes.NewReader(pdfData),
		false,
		w,
		r,
	)
}

func handleCWSWebhook(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.App.Channels().License() == nil || !*c.App.Channels().License().Features.Cloud {
		c.Err = model.NewAppError("Api4.handleCWSWebhook", "api.cloud.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Err = model.NewAppError("Api4.handleCWSWebhook", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var event *model.CWSWebhookPayload
	if err = json.Unmarshal(bodyBytes, &event); err != nil {
		c.Err = model.NewAppError("Api4.handleCWSWebhook", "api.cloud.app_error", nil, err.Error(), http.StatusInternalServerError)
		return
	}

	switch event.Event {
	case model.EventTypeFailedPayment:
		if nErr := c.App.SendPaymentFailedEmail(event.FailedPayment); nErr != nil {
			c.Err = nErr
			return
		}
	case model.EventTypeFailedPaymentNoCard:
		if nErr := c.App.SendNoCardPaymentFailedEmail(); nErr != nil {
			c.Err = nErr
			return
		}
	case model.EventTypeSendUpgradeConfirmationEmail:
		if nErr := c.App.SendUpgradeConfirmationEmail(); nErr != nil {
			c.Err = nErr
			return
		}
	case model.EventTypeSendAdminWelcomeEmail:
		user, appErr := c.App.GetUserByUsername(event.CloudWorkspaceOwner.UserName)
		if appErr != nil {
			c.Err = model.NewAppError("Api4.handleCWSWebhook", appErr.Id, nil, appErr.Error(), appErr.StatusCode)
			return
		}

		teams, appErr := c.App.GetAllTeams()
		if appErr != nil {
			c.Err = model.NewAppError("Api4.handleCWSWebhook", appErr.Id, nil, appErr.Error(), appErr.StatusCode)
			return
		}

		team := teams[0]

		subscription, err := c.App.Cloud().GetSubscription(user.Id)
		if err != nil {
			c.Err = model.NewAppError("Api4.handleCWSWebhook", "api.cloud.request_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := c.App.Srv().EmailService.SendCloudWelcomeEmail(user.Email, user.Locale, team.InviteId, subscription.GetWorkSpaceNameFromDNS(), subscription.DNS, *c.App.Config().ServiceSettings.SiteURL); err != nil {
			c.Err = model.NewAppError("SendCloudWelcomeEmail", "api.user.send_cloud_welcome_email.error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
	case model.EventTypeTrialWillEnd:
		endTimeStamp := event.SubscriptionTrialEndUnixTimeStamp
		t := time.Unix(endTimeStamp, 0)
		trialEndDate := fmt.Sprintf("%s %d, %d", t.Month(), t.Day(), t.Year())

		if appErr := c.App.SendCloudTrialEndWarningEmail(trialEndDate, *c.App.Config().ServiceSettings.SiteURL); appErr != nil {
			c.Err = appErr
			return
		}
	case model.EventTypeTrialEnded:
		if appErr := c.App.SendCloudTrialEndedEmail(); appErr != nil {
			c.Err = appErr
			return
		}
	case model.EventTypeSubscriptionChanged:
		// event.ProductLimits is nil if there was no change
		if event.ProductLimits != nil {
			if pluginsEnvironment := c.App.GetPluginsEnvironment(); pluginsEnvironment != nil {
				pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
					hooks.OnCloudLimitsUpdated(event.ProductLimits)
					return true
				}, plugin.OnCloudLimitsUpdatedID)
			}
			c.App.AdjustInProductLimits(event.ProductLimits, event.Subscription)
		}

		if err := c.App.Cloud().UpdateSubscriptionFromHook(event.ProductLimits, event.Subscription); err != nil {
			c.Err = model.NewAppError("Api4.handleCWSWebhook", "api.cloud.subscription.update_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		c.Logger.Info("Updated subscription from webhook event")

	default:
		c.Err = model.NewAppError("Api4.handleCWSWebhook", "api.cloud.cws_webhook_event_missing_error", nil, "", http.StatusNotFound)
		return
	}

	ReturnStatusOK(w)
}
