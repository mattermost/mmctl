// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"net/http"

	"github.com/mattermost/go-i18n/i18n"
	"github.com/pkg/errors"
	"github.com/throttled/throttled"
	"github.com/throttled/throttled/store/memstore"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/services/mailservice"
	"github.com/mattermost/mattermost-server/v5/utils"
)

const (
	emailRateLimitingMemstoreSize = 65536
	emailRateLimitingPerHour      = 20
	emailRateLimitingMaxBurst     = 20
)

func condenseSiteURL(siteURL string) string {
	parsedSiteURL, _ := url.Parse(siteURL)
	if parsedSiteURL.Path == "" || parsedSiteURL.Path == "/" {
		return parsedSiteURL.Host
	}

	return path.Join(parsedSiteURL.Host, parsedSiteURL.Path)
}

type EmailService struct {
	srv              *Server
	EmailRateLimiter *throttled.GCRARateLimiter
	EmailBatching    *EmailBatchingJob
}

func NewEmailService(srv *Server) (*EmailService, error) {
	service := &EmailService{srv: srv}
	if err := service.setupInviteEmailRateLimiting(); err != nil {
		return nil, err
	}
	service.InitEmailBatching()
	return service, nil
}

func (es *EmailService) setupInviteEmailRateLimiting() error {
	store, err := memstore.New(emailRateLimitingMemstoreSize)
	if err != nil {
		return errors.Wrap(err, "Unable to setup email rate limiting memstore.")
	}

	quota := throttled.RateQuota{
		MaxRate:  throttled.PerHour(emailRateLimitingPerHour),
		MaxBurst: emailRateLimitingMaxBurst,
	}

	rateLimiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil || rateLimiter == nil {
		return errors.Wrap(err, "Unable to setup email rate limiting GCRA rate limiter.")
	}

	es.EmailRateLimiter = rateLimiter
	return nil
}

func (es *EmailService) sendChangeUsernameEmail(oldUsername, newUsername, email, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.username_change_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName,
			"TeamDisplayName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("email_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.username_change_body.title")
	bodyPage.Props["Info"] = T("api.templates.username_change_body.info",
		map[string]interface{}{"TeamDisplayName": es.srv.Config().TeamSettings.SiteName, "NewUsername": newUsername})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("sendChangeUsernameEmail", "api.user.send_email_change_username_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) sendEmailChangeVerifyEmail(newUserEmail, locale, siteURL, token string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	link := fmt.Sprintf("%s/do_verify_email?token=%s&email=%s", siteURL, token, url.QueryEscape(newUserEmail))

	subject := T("api.templates.email_change_verify_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName,
			"TeamDisplayName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("email_change_verify_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.email_change_verify_body.title")
	bodyPage.Props["Info"] = T("api.templates.email_change_verify_body.info",
		map[string]interface{}{"TeamDisplayName": es.srv.Config().TeamSettings.SiteName})
	bodyPage.Props["VerifyUrl"] = link
	bodyPage.Props["VerifyButton"] = T("api.templates.email_change_verify_body.button")

	if err := es.sendMail(newUserEmail, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("sendEmailChangeVerifyEmail", "api.user.send_email_change_verify_email_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) sendEmailChangeEmail(oldEmail, newEmail, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.email_change_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName,
			"TeamDisplayName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("email_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.email_change_body.title")
	bodyPage.Props["Info"] = T("api.templates.email_change_body.info",
		map[string]interface{}{"TeamDisplayName": es.srv.Config().TeamSettings.SiteName, "NewEmail": newEmail})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := es.sendMail(oldEmail, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("sendEmailChangeEmail", "api.user.send_email_change_email_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) sendVerifyEmail(userEmail, locale, siteURL, token, redirect string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	link := fmt.Sprintf("%s/do_verify_email?token=%s&email=%s", siteURL, token, url.QueryEscape(userEmail))
	if redirect != "" {
		link += fmt.Sprintf("&redirect_to=%s", redirect)
	}

	serverURL := condenseSiteURL(siteURL)

	subject := T("api.templates.verify_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("verify_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.verify_body.title", map[string]interface{}{"ServerURL": serverURL})
	bodyPage.Props["Info"] = T("api.templates.verify_body.info")
	bodyPage.Props["VerifyUrl"] = link
	bodyPage.Props["Button"] = T("api.templates.verify_body.button")

	if err := es.sendMail(userEmail, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendVerifyEmail", "api.user.send_verify_email_and_forget.failed.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) SendSignInChangeEmail(email, method, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.signin_change_email.subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("signin_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.signin_change_email.body.title")
	bodyPage.Props["Info"] = T("api.templates.signin_change_email.body.info",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName, "Method": method})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendSignInChangeEmail", "api.user.send_sign_in_change_email_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) sendWelcomeEmail(userId string, email string, verified bool, locale, siteURL, redirect string) *model.AppError {
	if !*es.srv.Config().EmailSettings.SendEmailNotifications && !*es.srv.Config().EmailSettings.RequireEmailVerification {
		return model.NewAppError("SendWelcomeEmail", "api.user.send_welcome_email_and_forget.failed.error", nil, "Send Email Notifications and Require Email Verification is disabled in the system console", http.StatusInternalServerError)
	}

	T := utils.GetUserTranslations(locale)

	serverURL := condenseSiteURL(siteURL)

	subject := T("api.templates.welcome_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName,
			"ServerURL": serverURL})

	bodyPage := es.newEmailTemplate("welcome_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.welcome_body.title", map[string]interface{}{"ServerURL": serverURL})
	bodyPage.Props["Info"] = T("api.templates.welcome_body.info")
	bodyPage.Props["Button"] = T("api.templates.welcome_body.button")
	bodyPage.Props["Info2"] = T("api.templates.welcome_body.info2")
	bodyPage.Props["Info3"] = T("api.templates.welcome_body.info3")
	bodyPage.Props["SiteURL"] = siteURL

	if *es.srv.Config().NativeAppSettings.AppDownloadLink != "" {
		bodyPage.Props["AppDownloadInfo"] = T("api.templates.welcome_body.app_download_info")
		bodyPage.Props["AppDownloadLink"] = *es.srv.Config().NativeAppSettings.AppDownloadLink
	}

	if !verified && *es.srv.Config().EmailSettings.RequireEmailVerification {
		token, err := es.CreateVerifyEmailToken(userId, email)
		if err != nil {
			return err
		}
		link := fmt.Sprintf("%s/do_verify_email?token=%s&email=%s", siteURL, token.Token, url.QueryEscape(email))
		if redirect != "" {
			link += fmt.Sprintf("&redirect_to=%s", redirect)
		}
		bodyPage.Props["VerifyUrl"] = link
	}

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("sendWelcomeEmail", "api.user.send_welcome_email_and_forget.failed.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) sendPasswordChangeEmail(email, method, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.password_change_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName,
			"TeamDisplayName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("password_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.password_change_body.title")
	bodyPage.Props["Info"] = T("api.templates.password_change_body.info",
		map[string]interface{}{"TeamDisplayName": es.srv.Config().TeamSettings.SiteName, "TeamURL": siteURL, "Method": method})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("sendPasswordChangeEmail", "api.user.send_password_change_email_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) sendUserAccessTokenAddedEmail(email, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.user_access_token_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("password_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.user_access_token_body.title")
	bodyPage.Props["Info"] = T("api.templates.user_access_token_body.info",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName, "SiteURL": siteURL})
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("sendUserAccessTokenAddedEmail", "api.user.send_user_access_token.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) SendPasswordResetEmail(email string, token *model.Token, locale, siteURL string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	link := fmt.Sprintf("%s/reset_password_complete?token=%s", siteURL, url.QueryEscape(token.Token))

	subject := T("api.templates.reset_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("reset_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.reset_body.title")
	bodyPage.Props["Info1"] = utils.TranslateAsHtml(T, "api.templates.reset_body.info1", nil)
	bodyPage.Props["Info2"] = T("api.templates.reset_body.info2")
	bodyPage.Props["ResetUrl"] = link
	bodyPage.Props["Button"] = T("api.templates.reset_body.button")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendPasswordReset", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) sendMfaChangeEmail(email string, activated bool, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.mfa_change_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("mfa_change_body", locale)
	bodyPage.Props["SiteURL"] = siteURL

	if activated {
		bodyPage.Props["Info"] = T("api.templates.mfa_activated_body.info", map[string]interface{}{"SiteURL": siteURL})
		bodyPage.Props["Title"] = T("api.templates.mfa_activated_body.title")
	} else {
		bodyPage.Props["Info"] = T("api.templates.mfa_deactivated_body.info", map[string]interface{}{"SiteURL": siteURL})
		bodyPage.Props["Title"] = T("api.templates.mfa_deactivated_body.title")
	}
	bodyPage.Props["Warning"] = T("api.templates.email_warning")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendMfaChangeEmail", "api.user.send_mfa_change_email.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) SendInviteEmails(team *model.Team, senderName string, senderUserId string, invites []string, siteURL string) *model.AppError {
	if es.EmailRateLimiter == nil {
		return model.NewAppError("SendInviteEmails", "app.email.no_rate_limiter.app_error", nil, fmt.Sprintf("user_id=%s, team_id=%s", senderUserId, team.Id), http.StatusInternalServerError)
	}
	rateLimited, result, err := es.EmailRateLimiter.RateLimit(senderUserId, len(invites))
	if err != nil {
		return model.NewAppError("SendInviteEmails", "app.email.setup_rate_limiter.app_error", nil, fmt.Sprintf("user_id=%s, team_id=%s, error=%v", senderUserId, team.Id, err), http.StatusInternalServerError)
	}

	if rateLimited {
		return model.NewAppError("SendInviteEmails",
			"app.email.rate_limit_exceeded.app_error", map[string]interface{}{"RetryAfter": result.RetryAfter.String(), "ResetAfter": result.ResetAfter.String()},
			fmt.Sprintf("user_id=%s, team_id=%s, retry_after_secs=%f, reset_after_secs=%f",
				senderUserId, team.Id, result.RetryAfter.Seconds(), result.ResetAfter.Seconds()),
			http.StatusRequestEntityTooLarge)
	}

	for _, invite := range invites {
		if len(invite) > 0 {
			subject := utils.T("api.templates.invite_subject",
				map[string]interface{}{"SenderName": senderName,
					"TeamDisplayName": team.DisplayName,
					"SiteName":        es.srv.Config().TeamSettings.SiteName})

			bodyPage := es.newEmailTemplate("invite_body", "")
			bodyPage.Props["SiteURL"] = siteURL
			bodyPage.Props["Title"] = utils.T("api.templates.invite_body.title")
			bodyPage.Html["Info"] = utils.TranslateAsHtml(utils.T, "api.templates.invite_body.info",
				map[string]interface{}{"SenderName": senderName, "TeamDisplayName": team.DisplayName})
			bodyPage.Props["Button"] = utils.T("api.templates.invite_body.button")
			bodyPage.Html["ExtraInfo"] = utils.TranslateAsHtml(utils.T, "api.templates.invite_body.extra_info",
				map[string]interface{}{"TeamDisplayName": team.DisplayName})
			bodyPage.Props["TeamURL"] = siteURL + "/" + team.Name

			token := model.NewToken(
				TOKEN_TYPE_TEAM_INVITATION,
				model.MapToJson(map[string]string{"teamId": team.Id, "email": invite}),
			)

			props := make(map[string]string)
			props["email"] = invite
			props["display_name"] = team.DisplayName
			props["name"] = team.Name
			data := model.MapToJson(props)

			if err := es.srv.Store.Token().Save(token); err != nil {
				mlog.Error("Failed to send invite email successfully ", mlog.Err(err))
				continue
			}
			bodyPage.Props["Link"] = fmt.Sprintf("%s/signup_user_complete/?d=%s&t=%s", siteURL, url.QueryEscape(data), url.QueryEscape(token.Token))

			if err := es.sendMail(invite, subject, bodyPage.Render()); err != nil {
				mlog.Error("Failed to send invite email successfully ", mlog.Err(err))
			}
		}
	}
	return nil
}

func (es *EmailService) sendGuestInviteEmails(team *model.Team, channels []*model.Channel, senderName string, senderUserId string, senderProfileImage []byte, invites []string, siteURL string, message string) *model.AppError {
	if es.EmailRateLimiter == nil {
		return model.NewAppError("SendInviteEmails", "app.email.no_rate_limiter.app_error", nil, fmt.Sprintf("user_id=%s, team_id=%s", senderUserId, team.Id), http.StatusInternalServerError)
	}
	rateLimited, result, err := es.EmailRateLimiter.RateLimit(senderUserId, len(invites))
	if err != nil {
		return model.NewAppError("SendInviteEmails", "app.email.setup_rate_limiter.app_error", nil, fmt.Sprintf("user_id=%s, team_id=%s, error=%v", senderUserId, team.Id, err), http.StatusInternalServerError)
	}

	if rateLimited {
		return model.NewAppError("SendInviteEmails",
			"app.email.rate_limit_exceeded.app_error", map[string]interface{}{"RetryAfter": result.RetryAfter.String(), "ResetAfter": result.ResetAfter.String()},
			fmt.Sprintf("user_id=%s, team_id=%s, retry_after_secs=%f, reset_after_secs=%f",
				senderUserId, team.Id, result.RetryAfter.Seconds(), result.ResetAfter.Seconds()),
			http.StatusRequestEntityTooLarge)
	}

	for _, invite := range invites {
		if len(invite) > 0 {
			subject := utils.T("api.templates.invite_guest_subject",
				map[string]interface{}{"SenderName": senderName,
					"TeamDisplayName": team.DisplayName,
					"SiteName":        es.srv.Config().TeamSettings.SiteName})

			bodyPage := es.newEmailTemplate("invite_body", "")
			bodyPage.Props["SiteURL"] = siteURL
			bodyPage.Props["Title"] = utils.T("api.templates.invite_body.title")
			bodyPage.Html["Info"] = utils.TranslateAsHtml(utils.T, "api.templates.invite_body_guest.info",
				map[string]interface{}{"SenderName": senderName, "TeamDisplayName": team.DisplayName})
			bodyPage.Props["Button"] = utils.T("api.templates.invite_body.button")
			bodyPage.Props["SenderName"] = senderName
			bodyPage.Props["SenderId"] = senderUserId
			bodyPage.Props["Message"] = ""
			if message != "" {
				bodyPage.Props["Message"] = message
			}
			bodyPage.Html["ExtraInfo"] = utils.TranslateAsHtml(utils.T, "api.templates.invite_body.extra_info",
				map[string]interface{}{"TeamDisplayName": team.DisplayName})
			bodyPage.Props["TeamURL"] = siteURL + "/" + team.Name

			channelIds := []string{}
			for _, channel := range channels {
				channelIds = append(channelIds, channel.Id)
			}

			token := model.NewToken(
				TOKEN_TYPE_GUEST_INVITATION,
				model.MapToJson(map[string]string{
					"teamId":   team.Id,
					"channels": strings.Join(channelIds, " "),
					"email":    invite,
					"guest":    "true",
				}),
			)

			props := make(map[string]string)
			props["email"] = invite
			props["display_name"] = team.DisplayName
			props["name"] = team.Name
			data := model.MapToJson(props)

			if err := es.srv.Store.Token().Save(token); err != nil {
				mlog.Error("Failed to send invite email successfully ", mlog.Err(err))
				continue
			}
			bodyPage.Props["Link"] = fmt.Sprintf("%s/signup_user_complete/?d=%s&t=%s", siteURL, url.QueryEscape(data), url.QueryEscape(token.Token))

			if !*es.srv.Config().EmailSettings.SendEmailNotifications {
				mlog.Info("sending invitation ", mlog.String("to", invite), mlog.String("link", bodyPage.Props["Link"].(string)))
			}

			embeddedFiles := make(map[string]io.Reader)
			if message != "" {
				if senderProfileImage != nil {
					embeddedFiles = map[string]io.Reader{
						"user-avatar.png": bytes.NewReader(senderProfileImage),
					}
				}
			}

			if err := es.sendMailWithEmbeddedFiles(invite, subject, bodyPage.Render(), embeddedFiles); err != nil {
				mlog.Error("Failed to send invite email successfully", mlog.Err(err))
			}
		}
	}
	return nil
}

func (es *EmailService) newEmailTemplate(name, locale string) *utils.HTMLTemplate {
	t := utils.NewHTMLTemplate(es.srv.HTMLTemplates(), name)

	var localT i18n.TranslateFunc
	if locale != "" {
		localT = utils.GetUserTranslations(locale)
	} else {
		localT = utils.T
	}

	t.Props["Footer"] = localT("api.templates.email_footer")

	if *es.srv.Config().EmailSettings.FeedbackOrganization != "" {
		t.Props["Organization"] = localT("api.templates.email_organization") + *es.srv.Config().EmailSettings.FeedbackOrganization
	} else {
		t.Props["Organization"] = ""
	}

	t.Props["EmailInfo1"] = localT("api.templates.email_info1")
	t.Props["EmailInfo2"] = localT("api.templates.email_info2")
	t.Props["EmailInfo3"] = localT("api.templates.email_info3",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName})
	t.Props["SupportEmail"] = *es.srv.Config().SupportSettings.SupportEmail

	return t
}

func (es *EmailService) SendDeactivateAccountEmail(email string, locale, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	serverURL := condenseSiteURL(siteURL)

	subject := T("api.templates.deactivate_subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName,
			"ServerURL": serverURL})

	bodyPage := es.newEmailTemplate("deactivate_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.deactivate_body.title", map[string]interface{}{"ServerURL": serverURL})
	bodyPage.Props["Info"] = T("api.templates.deactivate_body.info",
		map[string]interface{}{"SiteURL": siteURL})
	bodyPage.Props["Warning"] = T("api.templates.deactivate_body.warning")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendDeactivateEmail", "api.user.send_deactivate_email_and_forget.failed.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) SendRemoveExpiredLicenseEmail(email string, locale, siteURL string, licenseId string) *model.AppError {
	T := utils.GetUserTranslations(locale)
	subject := T("api.templates.remove_expired_license.subject",
		map[string]interface{}{"SiteName": es.srv.Config().TeamSettings.SiteName})

	bodyPage := es.newEmailTemplate("remove_expired_license", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.remove_expired_license.body.title")
	bodyPage.Props["Link"] = fmt.Sprintf("%s?id=%s", model.LICENSE_RENEWAL_LINK, licenseId)
	bodyPage.Props["LinkButton"] = T("api.templates.remove_expired_license.body.renew_button")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendRemoveExpiredLicenseEmail", "api.license.remove_expired_license.failed.error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (es *EmailService) sendNotificationMail(to, subject, htmlBody string) *model.AppError {
	if !*es.srv.Config().EmailSettings.SendEmailNotifications {
		return nil
	}
	return es.sendMail(to, subject, htmlBody)
}

func (es *EmailService) sendMail(to, subject, htmlBody string) *model.AppError {
	return es.sendMailWithCC(to, subject, htmlBody, "")
}

func (es *EmailService) sendMailWithCC(to, subject, htmlBody string, ccMail string) *model.AppError {
	license := es.srv.License()
	return mailservice.SendMailUsingConfig(to, subject, htmlBody, es.srv.Config(), license != nil && *license.Features.Compliance, ccMail)
}

func (es *EmailService) sendMailWithEmbeddedFiles(to, subject, htmlBody string, embeddedFiles map[string]io.Reader) *model.AppError {
	license := es.srv.License()
	config := es.srv.Config()

	return mailservice.SendMailWithEmbeddedFilesUsingConfig(to, subject, htmlBody, embeddedFiles, config, license != nil && *license.Features.Compliance, "")
}

func (es *EmailService) CreateVerifyEmailToken(userId string, newEmail string) (*model.Token, *model.AppError) {
	tokenExtra := struct {
		UserId string
		Email  string
	}{
		userId,
		newEmail,
	}
	jsonData, err := json.Marshal(tokenExtra)

	if err != nil {
		return nil, model.NewAppError("CreateVerifyEmailToken", "api.user.create_email_token.error", nil, "", http.StatusInternalServerError)
	}

	token := model.NewToken(TOKEN_TYPE_VERIFY_EMAIL, string(jsonData))

	if err = es.srv.Store.Token().Save(token); err != nil {
		var appErr *model.AppError
		switch {
		case errors.As(err, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("CreateVerifyEmailToken", "app.recover.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return token, nil
}

func (es *EmailService) SendAtUserLimitWarningEmail(email string, locale string, siteURL string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.at_limit_subject")

	bodyPage := es.newEmailTemplate("reached_user_limit_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.at_limit_title")
	bodyPage.Props["Info1"] = T("api.templates.at_limit_info1")
	bodyPage.Props["Info2"] = T("api.templates.at_limit_info2")
	bodyPage.Props["Button"] = T("api.templates.upgrade_mattermost_cloud")
	bodyPage.Props["EmailUs"] = T("api.templates.email_us_anytime_at")

	bodyPage.Props["Footer"] = T("api.templates.copyright")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendAtUserLimitWarningEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) SendOverUserLimitWarningEmail(email string, locale string, siteURL string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.over_limit_subject")

	bodyPage := es.newEmailTemplate("reached_user_limit_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.over_limit_title")
	bodyPage.Props["Info1"] = T("api.templates.over_limit_info1")
	bodyPage.Props["Info2"] = T("api.templates.over_limit_info2")
	bodyPage.Props["Button"] = T("api.templates.upgrade_mattermost_cloud")
	bodyPage.Props["EmailUs"] = T("api.templates.email_us_anytime_at")

	bodyPage.Props["Footer"] = T("api.templates.copyright")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendOverUserLimitWarningEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) SendOverUserLimitThirtyDayWarningEmail(email string, locale string, siteURL string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.over_limit_30_days_subject")

	bodyPage := es.newEmailTemplate("over_user_limit_30_days_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.over_limit_30_days_title")
	bodyPage.Props["Info1"] = T("api.templates.over_limit_30_days_info1")
	bodyPage.Props["Info2"] = T("api.templates.over_limit_30_days_info2")
	bodyPage.Props["Info2Item1"] = T("api.templates.over_limit_30_days_info2_item1")
	bodyPage.Props["Info2Item2"] = T("api.templates.over_limit_30_days_info2_item2")
	bodyPage.Props["Info2Item3"] = T("api.templates.over_limit_30_days_info2_item3")
	bodyPage.Props["Button"] = T("api.templates.over_limit_fix_now")
	bodyPage.Props["EmailUs"] = T("api.templates.email_us_anytime_at")

	bodyPage.Props["Footer"] = T("api.templates.copyright")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendOverUserLimitWarningEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) SendOverUserLimitNinetyDayWarningEmail(email string, locale string, siteURL string, overLimitDate string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.over_limit_90_days_subject")

	bodyPage := es.newEmailTemplate("over_user_limit_90_days_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.over_limit_90_days_title")
	bodyPage.Props["Info1"] = T("api.templates.over_limit_90_days_info1", map[string]interface{}{"OverLimitDate": overLimitDate})
	bodyPage.Props["Info2"] = T("api.templates.over_limit_90_days_info2")
	bodyPage.Props["Info3"] = T("api.templates.over_limit_90_days_info3")
	bodyPage.Props["Info4"] = T("api.templates.over_limit_90_days_info4")
	bodyPage.Props["Button"] = T("api.templates.over_limit_fix_now")
	bodyPage.Props["EmailUs"] = T("api.templates.email_us_anytime_at")

	bodyPage.Props["Footer"] = T("api.templates.copyright")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendOverUserLimitWarningEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) SendOverUserLimitWorkspaceSuspendedWarningEmail(email string, locale string, siteURL string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.over_limit_suspended_subject")

	bodyPage := es.newEmailTemplate("over_user_limit_workspace_suspended_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.over_limit_suspended_title")
	bodyPage.Props["Info1"] = T("api.templates.over_limit_suspended_info1")
	bodyPage.Props["Info2"] = T("api.templates.over_limit_suspended_info2")
	bodyPage.Props["Button"] = T("api.templates.over_limit_suspended_contact_support")
	bodyPage.Props["EmailUs"] = T("api.templates.email_us_anytime_at")

	bodyPage.Props["Footer"] = T("api.templates.copyright")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendOverUserLimitWarningEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) SendOverUserFourteenDayWarningEmail(email string, locale string, siteURL string, overLimitDate string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.over_limit_14_days_subject")

	bodyPage := es.newEmailTemplate("over_user_limit_7_days_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.over_limit_14_days_title")
	bodyPage.Props["Info1"] = T("api.templates.over_limit_14_days_info1", map[string]interface{}{"OverLimitDate": overLimitDate})
	bodyPage.Props["Button"] = T("api.templates.over_limit_fix_now")
	bodyPage.Props["EmailUs"] = T("api.templates.email_us_anytime_at")

	bodyPage.Props["Footer"] = T("api.templates.copyright")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendOverUserLimitWarningEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) SendOverUserSevenDayWarningEmail(email string, locale string, siteURL string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.over_limit_7_days_subject")

	bodyPage := es.newEmailTemplate("over_user_limit_7_days_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.over_limit_7_days_title")
	bodyPage.Props["Info1"] = T("api.templates.over_limit_7_days_info1")
	bodyPage.Props["Button"] = T("api.templates.over_limit_fix_now")
	bodyPage.Props["EmailUs"] = T("api.templates.email_us_anytime_at")

	bodyPage.Props["Footer"] = T("api.templates.copyright")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendOverUserLimitWarningEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) SendSuspensionEmailToSupport(email string, installationID string, customerID string, subscriptionID string, siteURL string, userCount int64) (bool, *model.AppError) {
	// Localization not needed

	subject := fmt.Sprintf("Cloud Installation %s Scheduled Suspension", installationID)
	bodyPage := es.newEmailTemplate("over_user_limit_support_body", "en")
	bodyPage.Props["CustomerID"] = customerID
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["SubscriptionID"] = subscriptionID
	bodyPage.Props["InstallationID"] = installationID
	bodyPage.Props["SuspensionDate"] = time.Now().AddDate(0, 0, 61).Format("2006-01-02")
	bodyPage.Props["UserCount"] = userCount

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendOverUserLimitWarningEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) SendPaymentFailedEmail(email string, locale string, failedPayment *model.FailedPayment, siteURL string) (bool, *model.AppError) {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.payment_failed.subject")

	bodyPage := es.newEmailTemplate("payment_failed_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.payment_failed.title")
	bodyPage.Props["Info1"] = T("api.templates.payment_failed.info1", map[string]interface{}{"CardBrand": failedPayment.CardBrand, "LastFour": failedPayment.LastFour})
	bodyPage.Props["Info2"] = T("api.templates.payment_failed.info2")
	bodyPage.Props["Info3"] = T("api.templates.payment_failed.info3")
	bodyPage.Props["Button"] = T("api.templates.over_limit_fix_now")
	bodyPage.Props["EmailUs"] = T("api.templates.email_us_anytime_at")

	bodyPage.Props["FailedReason"] = failedPayment.FailureMessage

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return false, model.NewAppError("SendPaymentFailedEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return true, nil
}

func (es *EmailService) SendNoCardPaymentFailedEmail(email string, locale string, siteURL string) *model.AppError {
	T := utils.GetUserTranslations(locale)

	subject := T("api.templates.payment_failed_no_card.subject")

	bodyPage := es.newEmailTemplate("payment_failed_no_card_body", locale)
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = T("api.templates.payment_failed_no_card.title")
	bodyPage.Props["Info1"] = T("api.templates.payment_failed_no_card.info1")
	bodyPage.Props["Info3"] = T("api.templates.payment_failed_no_card.info3")
	bodyPage.Props["Button"] = T("api.templates.payment_failed_no_card.button")
	bodyPage.Props["EmailUs"] = T("api.templates.email_us_anytime_at")

	if err := es.sendMail(email, subject, bodyPage.Render()); err != nil {
		return model.NewAppError("SendPaymentFailedEmail", "api.user.send_password_reset.send.app_error", nil, "err="+err.Message, http.StatusInternalServerError)
	}

	return nil
}
