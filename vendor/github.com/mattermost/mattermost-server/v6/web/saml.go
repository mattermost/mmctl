// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package web

import (
	b64 "encoding/base64"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v6/audit"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"github.com/mattermost/mattermost-server/v6/utils"
)

func (w *Web) InitSaml() {
	w.MainRouter.Handle("/login/sso/saml", w.APIHandler(loginWithSaml)).Methods("GET")
	w.MainRouter.Handle("/login/sso/saml", w.APIHandlerTrustRequester(completeSaml)).Methods("POST")
}

func loginWithSaml(c *Context, w http.ResponseWriter, r *http.Request) {
	samlInterface := c.App.Saml()

	if samlInterface == nil {
		c.Err = model.NewAppError("loginWithSaml", "api.user.saml.not_available.app_error", nil, "", http.StatusFound)
		return
	}

	teamId, err := c.App.GetTeamIdFromQuery(r.URL.Query())
	if err != nil {
		c.Err = err
		return
	}
	action := r.URL.Query().Get("action")
	isMobile := action == model.OAuthActionMobile
	redirectURL := html.EscapeString(r.URL.Query().Get("redirect_to"))
	relayProps := map[string]string{}
	relayState := ""

	if action != "" {
		relayProps["team_id"] = teamId
		relayProps["action"] = action
		if action == model.OAuthActionEmailToSSO {
			relayProps["email"] = r.URL.Query().Get("email")
		}
	}

	if redirectURL != "" {
		if isMobile && !utils.IsValidMobileAuthRedirectURL(c.App.Config(), redirectURL) {
			invalidSchemeErr := model.NewAppError("loginWithOAuth", "api.invalid_custom_url_scheme", nil, "", http.StatusBadRequest)
			utils.RenderMobileError(c.App.Config(), w, invalidSchemeErr, redirectURL)
			return
		}
		relayProps["redirect_to"] = redirectURL
	}

	relayProps[model.UserAuthServiceIsMobile] = strconv.FormatBool(isMobile)

	if len(relayProps) > 0 {
		relayState = b64.StdEncoding.EncodeToString([]byte(model.MapToJSON(relayProps)))
	}

	data, err := samlInterface.BuildRequest(relayState)
	if err != nil {
		c.Err = err
		return
	}
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	http.Redirect(w, r, data.URL, http.StatusFound)
}

func completeSaml(c *Context, w http.ResponseWriter, r *http.Request) {
	samlInterface := c.App.Saml()

	if samlInterface == nil {
		c.Err = model.NewAppError("completeSaml", "api.user.saml.not_available.app_error", nil, "", http.StatusFound)
		return
	}

	//Validate that the user is with SAML and all that
	encodedXML := r.FormValue("SAMLResponse")
	relayState := r.FormValue("RelayState")

	relayProps := make(map[string]string)
	if relayState != "" {
		stateStr := ""
		b, err := b64.StdEncoding.DecodeString(relayState)
		if err != nil {
			c.Err = model.NewAppError("completeSaml", "api.user.authorize_oauth_user.invalid_state.app_error", nil, err.Error(), http.StatusFound)
			return
		}
		stateStr = string(b)
		relayProps = model.MapFromJSON(strings.NewReader(stateStr))
	}

	auditRec := c.MakeAuditRecord("completeSaml", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	action := relayProps["action"]
	auditRec.AddMeta("action", action)

	isMobile := action == model.OAuthActionMobile
	redirectURL := ""
	hasRedirectURL := false
	if val, ok := relayProps["redirect_to"]; ok {
		redirectURL = val
		hasRedirectURL = val != ""
	}
	redirectURL = fullyQualifiedRedirectURL(c.GetSiteURLHeader(), redirectURL)

	handleError := func(err *model.AppError) {
		if isMobile && hasRedirectURL {
			err.Translate(c.AppContext.T)
			utils.RenderMobileError(c.App.Config(), w, err, redirectURL)
		} else {
			c.Err = err
			c.Err.StatusCode = http.StatusFound
		}
	}

	user, err := samlInterface.DoLogin(c.AppContext, encodedXML, relayProps)
	if err != nil {
		c.LogAudit("fail")
		mlog.Error(err.Error())
		handleError(err)
		return
	}

	if err = c.App.CheckUserAllAuthenticationCriteria(user, ""); err != nil {
		mlog.Error(err.Error())
		handleError(err)
		return
	}

	switch action {
	case model.OAuthActionSignup:
		if teamId := relayProps["team_id"]; teamId != "" {
			if err = c.App.AddUserToTeamByTeamId(c.AppContext, teamId, user); err != nil {
				c.LogErrorByCode(err)
				break
			}
			c.App.AddDirectChannels(teamId, user)
		}
	case model.OAuthActionEmailToSSO:
		if err = c.App.RevokeAllSessions(user.Id); err != nil {
			c.Err = err
			return
		}
		auditRec.AddMeta("revoked_user_id", user.Id)
		auditRec.AddMeta("revoked", "Revoked all sessions for user")

		c.LogAuditWithUserId(user.Id, "Revoked all sessions for user")
		c.App.Srv().Go(func() {
			if err := c.App.Srv().EmailService.SendSignInChangeEmail(user.Email, strings.Title(model.UserAuthServiceSaml)+" SSO", user.Locale, c.App.GetSiteURL()); err != nil {
				c.LogErrorByCode(model.NewAppError("SendSignInChangeEmail", "api.user.send_sign_in_change_email_and_forget.error", nil, err.Error(), http.StatusInternalServerError))
			}
		})
	}

	auditRec.AddMeta("obtained_user_id", user.Id)
	c.LogAuditWithUserId(user.Id, "obtained user")

	err = c.App.DoLogin(c.AppContext, w, r, user, "", isMobile, false, true)
	if err != nil {
		mlog.Error(err.Error())
		handleError(err)
		return
	}

	auditRec.Success()
	c.LogAuditWithUserId(user.Id, "success")

	c.App.AttachSessionCookies(c.AppContext, w, r)

	if hasRedirectURL {
		if isMobile {
			// Mobile clients with redirect url support
			redirectURL = utils.AppendQueryParamsToURL(redirectURL, map[string]string{
				model.SessionCookieToken: c.AppContext.Session().Token,
				model.SessionCookieCsrf:  c.AppContext.Session().GetCSRF(),
			})
			utils.RenderMobileAuthComplete(w, redirectURL)
		} else {
			http.Redirect(w, r, redirectURL, http.StatusFound)
		}
		return
	}

	switch action {
	// Mobile clients with web view implementation
	case model.OAuthActionMobile:
		ReturnStatusOK(w)
	case model.OAuthActionEmailToSSO:
		http.Redirect(w, r, c.GetSiteURLHeader()+"/login?extra=signin_change", http.StatusFound)
	default:
		http.Redirect(w, r, c.GetSiteURLHeader(), http.StatusFound)
	}
}
