// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package web

import (
	b64 "encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/audit"
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (w *Web) InitSaml() {
	w.MainRouter.Handle("/login/sso/saml", w.ApiHandler(loginWithSaml)).Methods("GET")
	w.MainRouter.Handle("/login/sso/saml", w.ApiHandlerTrustRequester(completeSaml)).Methods("POST")
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
	isMobile := action == model.OAUTH_ACTION_MOBILE
	redirectTo := r.URL.Query().Get("redirect_to")
	relayProps := map[string]string{}
	relayState := ""

	if len(action) != 0 {
		relayProps["team_id"] = teamId
		relayProps["action"] = action
		if action == model.OAUTH_ACTION_EMAIL_TO_SSO {
			relayProps["email"] = r.URL.Query().Get("email")
		}
	}

	if len(redirectTo) != 0 {
		relayProps["redirect_to"] = redirectTo
	}

	relayProps[model.USER_AUTH_SERVICE_IS_MOBILE] = strconv.FormatBool(isMobile)

	if len(relayProps) > 0 {
		relayState = b64.StdEncoding.EncodeToString([]byte(model.MapToJson(relayProps)))
	}

	if data, err := samlInterface.BuildRequest(relayState); err != nil {
		c.Err = err
		return
	} else {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		http.Redirect(w, r, data.URL, http.StatusFound)
	}
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
	if len(relayState) > 0 {
		stateStr := ""
		if b, err := b64.StdEncoding.DecodeString(relayState); err != nil {
			c.Err = model.NewAppError("completeSaml", "api.user.authorize_oauth_user.invalid_state.app_error", nil, err.Error(), http.StatusFound)
			return
		} else {
			stateStr = string(b)
		}
		relayProps = model.MapFromJson(strings.NewReader(stateStr))
	}

	auditRec := c.MakeAuditRecord("completeSaml", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	action := relayProps["action"]
	auditRec.AddMeta("action", action)

	user, err := samlInterface.DoLogin(encodedXML, relayProps)
	if err != nil {
		c.LogAudit("fail")
		if action == model.OAUTH_ACTION_MOBILE {
			err.Translate(c.App.T)
			w.Write([]byte(err.ToJson()))
		} else {
			c.Err = err
			c.Err.StatusCode = http.StatusFound
		}
		return
	}

	if err = c.App.CheckUserAllAuthenticationCriteria(user, ""); err != nil {
		c.Err = err
		c.Err.StatusCode = http.StatusFound
		return
	}

	switch action {
	case model.OAUTH_ACTION_SIGNUP:
		if teamId := relayProps["team_id"]; teamId != "" {
			if err = c.App.AddUserToTeamByTeamId(teamId, user); err != nil {
				mlog.Error(err.Error())
				break
			}
			c.App.AddDirectChannels(teamId, user)
		}
	case model.OAUTH_ACTION_EMAIL_TO_SSO:
		if err = c.App.RevokeAllSessions(user.Id); err != nil {
			c.Err = err
			return
		}
		auditRec.AddMeta("revoked_user_id", user.Id)
		auditRec.AddMeta("revoked", "Revoked all sessions for user")

		c.LogAuditWithUserId(user.Id, "Revoked all sessions for user")
		c.App.Srv().Go(func() {
			if err = c.App.Srv().EmailService.SendSignInChangeEmail(user.Email, strings.Title(model.USER_AUTH_SERVICE_SAML)+" SSO", user.Locale, c.App.GetSiteURL()); err != nil {
				mlog.Error(err.Error())
			}
		})
	}

	auditRec.AddMeta("obtained_user_id", user.Id)
	c.LogAuditWithUserId(user.Id, "obtained user")

	isMobile, parseErr := strconv.ParseBool(relayProps[model.USER_AUTH_SERVICE_IS_MOBILE])
	if parseErr != nil {
		mlog.Error("Error parsing boolean property from relay props", mlog.Err(parseErr))
	}
	err = c.App.DoLogin(w, r, user, "", isMobile, false, true)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	c.LogAuditWithUserId(user.Id, "success")

	c.App.AttachSessionCookies(w, r)

	if val, ok := relayProps["redirect_to"]; ok {
		http.Redirect(w, r, c.GetSiteURLHeader()+val, http.StatusFound)
		return
	}

	switch action {
	case model.OAUTH_ACTION_MOBILE:
		ReturnStatusOK(w)
	case model.OAUTH_ACTION_EMAIL_TO_SSO:
		http.Redirect(w, r, c.GetSiteURLHeader()+"/login?extra=signin_change", http.StatusFound)
	default:
		http.Redirect(w, r, c.GetSiteURLHeader(), http.StatusFound)
	}
}
