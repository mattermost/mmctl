// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package platform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/einterfaces"
	"github.com/mattermost/mattermost-server/v6/jobs"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"github.com/mattermost/mattermost-server/v6/utils"
)

const (
	LicenseEnv                = "MM_LICENSE"
	LicenseRenewalURL         = "https://customers.mattermost.com/subscribe/renew"
	JWTDefaultTokenExpiration = 7 * 24 * time.Hour // 7 days of expiration
)

var (
	RequestTrialURL = "https://customers.mattermost.com/api/v1/trials"
)

// JWTClaims custom JWT claims with the needed information for the
// renewal process
type JWTClaims struct {
	LicenseID   string `json:"license_id"`
	ActiveUsers int64  `json:"active_users"`
	jwt.StandardClaims
}

func (ps *PlatformService) LicenseManager() einterfaces.LicenseInterface {
	return ps.licenseManager
}

func (ps *PlatformService) SetLicenseManager(impl einterfaces.LicenseInterface) {
	ps.licenseManager = impl
}

func (ps *PlatformService) License() *model.License {
	license, _ := ps.licenseValue.Load().(*model.License)
	return license
}

func (ps *PlatformService) LoadLicense() {
	// ENV var overrides all other sources of license.
	licenseStr := os.Getenv(LicenseEnv)
	if licenseStr != "" {
		license, err := utils.LicenseValidator.LicenseFromBytes([]byte(licenseStr))
		if err != nil {
			ps.logger.Error("Failed to read license set in environment.", mlog.Err(err))
			return
		}

		// skip the restrictions if license is a sanctioned trial
		if !license.IsSanctionedTrial() && license.IsTrialLicense() {
			canStartTrialLicense, err := ps.licenseManager.CanStartTrial()
			if err != nil {
				ps.logger.Error("Failed to validate trial eligibility.", mlog.Err(err))
				return
			}

			if !canStartTrialLicense {
				ps.logger.Info("Cannot start trial multiple times.")
				return
			}
		}

		if ps.ValidateAndSetLicenseBytes([]byte(licenseStr)) {
			ps.logger.Info("License key from ENV is valid, unlocking enterprise features.")
		}
		return
	}

	licenseId := ""
	props, nErr := ps.Store.System().Get()
	if nErr == nil {
		licenseId = props[model.SystemActiveLicenseId]
	}

	if !model.IsValidId(licenseId) {
		// Lets attempt to load the file from disk since it was missing from the DB
		license, licenseBytes := utils.GetAndValidateLicenseFileFromDisk(*ps.Config().ServiceSettings.LicenseFileLocation)

		if license != nil {
			if _, err := ps.SaveLicense(licenseBytes); err != nil {
				ps.logger.Error("Failed to save license key loaded from disk.", mlog.Err(err))
			} else {
				licenseId = license.Id
			}
		}
	}

	record, nErr := ps.Store.License().Get(licenseId)
	if nErr != nil {
		ps.logger.Error("License key from https://mattermost.com required to unlock enterprise features.", mlog.Err(nErr))
		ps.SetLicense(nil)
		return
	}

	ps.ValidateAndSetLicenseBytes([]byte(record.Bytes))
	ps.logger.Info("License key valid unlocking enterprise features.")
}

func (ps *PlatformService) SaveLicense(licenseBytes []byte) (*model.License, *model.AppError) {
	success, licenseStr := utils.LicenseValidator.ValidateLicense(licenseBytes)
	if !success {
		return nil, model.NewAppError("addLicense", model.InvalidLicenseError, nil, "", http.StatusBadRequest)
	}

	var license model.License
	if jsonErr := json.Unmarshal([]byte(licenseStr), &license); jsonErr != nil {
		return nil, model.NewAppError("addLicense", "api.unmarshal_error", nil, "", http.StatusInternalServerError).Wrap(jsonErr)
	}

	uniqueUserCount, err := ps.Store.User().Count(model.UserCountOptions{})
	if err != nil {
		return nil, model.NewAppError("addLicense", "api.license.add_license.invalid_count.app_error", nil, "", http.StatusBadRequest).Wrap(err)
	}

	if uniqueUserCount > int64(*license.Features.Users) {
		return nil, model.NewAppError("addLicense", "api.license.add_license.unique_users.app_error", map[string]any{"Users": *license.Features.Users, "Count": uniqueUserCount}, "", http.StatusBadRequest)
	}

	if license.IsExpired() {
		return nil, model.NewAppError("addLicense", model.ExpiredLicenseError, nil, "", http.StatusBadRequest)
	}

	if *ps.Config().JobSettings.RunJobs && ps.Jobs != nil {
		if err := ps.Jobs.StopWorkers(); err != nil && !errors.Is(err, jobs.ErrWorkersNotRunning) {
			ps.logger.Warn("Stopping job server workers failed", mlog.Err(err))
		}
	}

	if *ps.Config().JobSettings.RunScheduler && ps.Jobs != nil {
		if err := ps.Jobs.StopSchedulers(); err != nil && !errors.Is(err, jobs.ErrSchedulersNotRunning) {
			ps.logger.Error("Stopping job server schedulers failed", mlog.Err(err))
		}
	}

	defer func() {
		// restart job server workers - this handles the edge case where a license file is uploaded, but the job server
		// doesn't start until the server is restarted, which prevents the 'run job now' buttons in system console from
		// functioning as expected
		if *ps.Config().JobSettings.RunJobs && ps.Jobs != nil {
			if err := ps.Jobs.StartWorkers(); err != nil {
				ps.logger.Error("Starting job server workers failed", mlog.Err(err))
			}
		}
		if *ps.Config().JobSettings.RunScheduler && ps.Jobs != nil {
			if err := ps.Jobs.StartSchedulers(); err != nil && !errors.Is(err, jobs.ErrSchedulersRunning) {
				ps.logger.Error("Starting job server schedulers failed", mlog.Err(err))
			}
		}
	}()

	if ok := ps.SetLicense(&license); !ok {
		return nil, model.NewAppError("addLicense", model.ExpiredLicenseError, nil, "", http.StatusBadRequest)
	}

	record := &model.LicenseRecord{}
	record.Id = license.Id
	record.Bytes = string(licenseBytes)

	_, nErr := ps.Store.License().Save(record)
	if nErr != nil {
		ps.RemoveLicense()
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("addLicense", "api.license.add_license.save.app_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
		}
	}

	sysVar := &model.System{}
	sysVar.Name = model.SystemActiveLicenseId
	sysVar.Value = license.Id
	if err := ps.Store.System().SaveOrUpdate(sysVar); err != nil {
		ps.RemoveLicense()
		return nil, model.NewAppError("addLicense", "api.license.add_license.save_active.app_error", nil, "", http.StatusInternalServerError)
	}

	ps.ReloadConfig()
	ps.InvalidateAllCaches()

	return &license, nil
}

func (ps *PlatformService) SetLicense(license *model.License) bool {
	oldLicense := ps.licenseValue.Load()

	defer func() {
		for _, listener := range ps.licenseListeners {
			if oldLicense == nil {
				listener(nil, license)
			} else {
				listener(oldLicense.(*model.License), license)
			}
		}
	}()

	if license != nil {
		license.Features.SetDefaults()

		ps.licenseValue.Store(license)

		ps.clientLicenseValue.Store(utils.GetClientLicense(license))
		return true
	}

	ps.licenseValue.Store((*model.License)(nil))
	ps.clientLicenseValue.Store(map[string]string(nil))

	return false
}

func (ps *PlatformService) ValidateAndSetLicenseBytes(b []byte) bool {
	if success, licenseStr := utils.LicenseValidator.ValidateLicense(b); success {
		var license model.License
		if jsonErr := json.Unmarshal([]byte(licenseStr), &license); jsonErr != nil {
			ps.logger.Warn("Failed to decode license from JSON", mlog.Err(jsonErr))
			return false
		}
		ps.SetLicense(&license)
		return true
	}

	ps.logger.Warn("No valid enterprise license found")
	return false
}

func (ps *PlatformService) SetClientLicense(m map[string]string) {
	ps.clientLicenseValue.Store(m)
}

func (ps *PlatformService) ClientLicense() map[string]string {
	if clientLicense, _ := ps.clientLicenseValue.Load().(map[string]string); clientLicense != nil {
		return clientLicense
	}
	return map[string]string{"IsLicensed": "false"}
}

func (ps *PlatformService) RemoveLicense() *model.AppError {
	if license, _ := ps.licenseValue.Load().(*model.License); license == nil {
		return nil
	}

	ps.logger.Info("Remove license.", mlog.String("id", model.SystemActiveLicenseId))

	sysVar := &model.System{}
	sysVar.Name = model.SystemActiveLicenseId
	sysVar.Value = ""

	if err := ps.Store.System().SaveOrUpdate(sysVar); err != nil {
		return model.NewAppError("RemoveLicense", "app.system.save.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	ps.SetLicense(nil)
	ps.ReloadConfig()
	ps.InvalidateAllCaches()

	return nil
}

func (ps *PlatformService) AddLicenseListener(listener func(oldLicense, newLicense *model.License)) string {
	id := model.NewId()
	ps.licenseListeners[id] = listener
	return id
}

func (ps *PlatformService) RemoveLicenseListener(id string) {
	delete(ps.licenseListeners, id)
}

func (ps *PlatformService) GetSanitizedClientLicense() map[string]string {
	return utils.GetSanitizedClientLicense(ps.ClientLicense())
}

// RequestTrialLicense request a trial license from the mattermost official license server
func (ps *PlatformService) RequestTrialLicense(trialRequest *model.TrialLicenseRequest) *model.AppError {
	trialRequestJSON, err := json.Marshal(trialRequest)
	if err != nil {
		return model.NewAppError("RequestTrialLicense", "api.unmarshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	resp, err := http.Post(RequestTrialURL, "application/json", bytes.NewBuffer(trialRequestJSON))
	if err != nil {
		return model.NewAppError("RequestTrialLicense", "api.license.request_trial_license.app_error", nil, "", http.StatusBadRequest).Wrap(err)
	}
	defer resp.Body.Close()

	// CloudFlare sitting in front of the Customer Portal will block this request with a 451 response code in the event that the request originates from a country sanctioned by the U.S. Government.
	if resp.StatusCode == http.StatusUnavailableForLegalReasons {
		return model.NewAppError("RequestTrialLicense", "api.license.request_trial_license.embargoed", nil, "Request for trial license came from an embargoed country", http.StatusUnavailableForLegalReasons)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.NewAppError("RequestTrialLicense", "api.license.request_trial_license.app_error", nil,
			fmt.Sprintf("Unexpected HTTP status code %q returned by server", resp.Status), http.StatusInternalServerError)
	}

	var licenseResponse map[string]string
	err = json.NewDecoder(resp.Body).Decode(&licenseResponse)
	if err != nil {
		ps.logger.Warn("Error decoding license response", mlog.Err(err))
	}

	if _, ok := licenseResponse["license"]; !ok {
		return model.NewAppError("RequestTrialLicense", "api.license.request_trial_license.app_error", nil, licenseResponse["message"], http.StatusBadRequest)
	}

	if _, err := ps.SaveLicense([]byte(licenseResponse["license"])); err != nil {
		return err
	}

	ps.ReloadConfig()
	ps.InvalidateAllCaches()

	return nil
}

// GenerateRenewalToken returns a renewal token that expires after duration expiration
func (ps *PlatformService) GenerateRenewalToken(expiration time.Duration) (string, *model.AppError) {
	license := ps.License()
	if license == nil {
		return "", model.NewAppError("GenerateRenewalToken", "app.license.generate_renewal_token.no_license", nil, "", http.StatusBadRequest)
	}

	if license.IsCloud() {
		return "", model.NewAppError("GenerateRenewalToken", "app.license.generate_renewal_token.bad_license", nil, "", http.StatusBadRequest)
	}

	activeUsers, err := ps.Store.User().Count(model.UserCountOptions{})
	if err != nil {
		return "", model.NewAppError("GenerateRenewalToken", "app.license.generate_renewal_token.app_error",
			nil, "", http.StatusInternalServerError).Wrap(err)
	}

	expirationTime := time.Now().UTC().Add(expiration)
	claims := &JWTClaims{
		LicenseID:   license.Id,
		ActiveUsers: activeUsers,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(license.Customer.Email))
	if err != nil {
		return "", model.NewAppError("GenerateRenewalToken", "app.license.generate_renewal_token.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	return tokenString, nil
}

// GenerateLicenseRenewalLink returns a link that points to the CWS where clients can renew license
func (ps *PlatformService) GenerateLicenseRenewalLink() (string, string, *model.AppError) {
	renewalToken, err := ps.GenerateRenewalToken(JWTDefaultTokenExpiration)
	if err != nil {
		return "", "", err
	}
	renewalLink := LicenseRenewalURL + "?token=" + renewalToken
	return renewalLink, renewalToken, nil
}
