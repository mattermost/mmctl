// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/mattermost/mattermost-server/v6/audit"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

func (api *API) InitJob() {
	api.BaseRoutes.Jobs.Handle("", api.APISessionRequired(getJobs)).Methods("GET")
	api.BaseRoutes.Jobs.Handle("", api.APISessionRequired(createJob)).Methods("POST")
	api.BaseRoutes.Jobs.Handle("/{job_id:[A-Za-z0-9]+}", api.APISessionRequired(getJob)).Methods("GET")
	api.BaseRoutes.Jobs.Handle("/{job_id:[A-Za-z0-9]+}/download", api.APISessionRequiredTrustRequester(downloadJob)).Methods("GET")
	api.BaseRoutes.Jobs.Handle("/{job_id:[A-Za-z0-9]+}/cancel", api.APISessionRequired(cancelJob)).Methods("POST")
	api.BaseRoutes.Jobs.Handle("/type/{job_type:[A-Za-z0-9_-]+}", api.APISessionRequired(getJobsByType)).Methods("GET")
}

func getJob(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireJobId()
	if c.Err != nil {
		return
	}

	job, err := c.App.GetJob(c.Params.JobId)
	if err != nil {
		c.Err = err
		return
	}

	hasPermission, permissionRequired := c.App.SessionHasPermissionToReadJob(*c.AppContext.Session(), job.Type)
	if permissionRequired == nil {
		c.Err = model.NewAppError("getJob", "api.job.retrieve.nopermissions", nil, "", http.StatusBadRequest)
		return
	}
	if !hasPermission {
		c.SetPermissionError(permissionRequired)
		return
	}

	if err := json.NewEncoder(w).Encode(job); err != nil {
		c.Logger.Warn("Error while writing response", mlog.Err(err))
	}
}

func downloadJob(c *Context, w http.ResponseWriter, r *http.Request) {
	config := c.App.Config()
	const FilePath = "export"
	const FileMime = "application/zip"

	c.RequireJobId()
	if c.Err != nil {
		return
	}

	if !*config.MessageExportSettings.DownloadExportResults {
		c.Err = model.NewAppError("downloadExportResultsNotEnabled", "app.job.download_export_results_not_enabled", nil, "", http.StatusNotImplemented)
		return
	}

	job, err := c.App.GetJob(c.Params.JobId)
	if err != nil {
		c.Err = err
		return
	}

	// Currently, this endpoint only supports downloading the compliance report.
	// If you need to download another job type, you will need to alter this section of the code to accommodate it.
	if job.Type == model.JobTypeMessageExport && !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionDownloadComplianceExportResult) {
		c.SetPermissionError(model.PermissionDownloadComplianceExportResult)
		return
	} else if job.Type != model.JobTypeMessageExport {
		c.Err = model.NewAppError("unableToDownloadJob", "api.job.unable_to_download_job.incorrect_job_type", nil, "", http.StatusBadRequest)
		return
	}

	isDownloadable, _ := strconv.ParseBool(job.Data["is_downloadable"])
	if !isDownloadable {
		c.Err = model.NewAppError("unableToDownloadJob", "api.job.unable_to_download_job", nil, "", http.StatusBadRequest)
		return
	}

	fileName := job.Id + ".zip"
	filePath := filepath.Join(FilePath, fileName)
	fileReader, err := c.App.FileReader(filePath)
	if err != nil {
		c.Err = err
		c.Err.StatusCode = http.StatusNotFound
		return
	}
	defer fileReader.Close()

	// We are able to pass 0 for content size due to the fact that Golang's serveContent (https://golang.org/src/net/http/fs.go)
	// already sets that for us
	writeFileResponse(fileName, FileMime, 0, time.Unix(0, job.LastActivityAt*int64(1000*1000)), *c.App.Config().ServiceSettings.WebserverMode, fileReader, true, w, r)
}

func createJob(c *Context, w http.ResponseWriter, r *http.Request) {
	var job model.Job
	if jsonErr := json.NewDecoder(r.Body).Decode(&job); jsonErr != nil {
		c.SetInvalidParamWithErr("job", jsonErr)
		return
	}

	auditRec := c.MakeAuditRecord("createJob", audit.Fail)
	defer c.LogAuditRec(auditRec)
	auditRec.AddEventParameter("job", job)

	hasPermission, permissionRequired := c.App.SessionHasPermissionToCreateJob(*c.AppContext.Session(), &job)
	if permissionRequired == nil {
		c.Err = model.NewAppError("unableToCreateJob", "api.job.unable_to_create_job.incorrect_job_type", nil, "", http.StatusBadRequest)
		return
	}

	if !hasPermission {
		c.SetPermissionError(permissionRequired)
		return
	}

	rjob, err := c.App.CreateJob(&job)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	auditRec.AddEventResultState(rjob)
	auditRec.AddEventObjectType("job")

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(rjob); err != nil {
		c.Logger.Warn("Error while writing response", mlog.Err(err))
	}
}

func getJobs(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.Err != nil {
		return
	}

	var validJobTypes []string
	for _, jobType := range model.AllJobTypes {
		hasPermission, permissionRequired := c.App.SessionHasPermissionToReadJob(*c.AppContext.Session(), jobType)
		if permissionRequired == nil {
			mlog.Warn("The job types of a job you are trying to retrieve does not contain permissions", mlog.String("jobType", jobType))
			continue
		}
		if hasPermission {
			validJobTypes = append(validJobTypes, jobType)
		}
	}
	if len(validJobTypes) == 0 {
		c.SetPermissionError()
		return
	}

	jobs, appErr := c.App.GetJobsByTypesPage(validJobTypes, c.Params.Page, c.Params.PerPage)
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(jobs)
	if err != nil {
		c.Err = model.NewAppError("getJobs", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}
	w.Write(js)
}

func getJobsByType(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireJobType()
	if c.Err != nil {
		return
	}

	hasPermission, permissionRequired := c.App.SessionHasPermissionToReadJob(*c.AppContext.Session(), c.Params.JobType)
	if permissionRequired == nil {
		c.Err = model.NewAppError("getJobsByType", "api.job.retrieve.nopermissions", nil, "", http.StatusBadRequest)
		return
	}
	if !hasPermission {
		c.SetPermissionError(permissionRequired)
		return
	}

	jobs, appErr := c.App.GetJobsByTypePage(c.Params.JobType, c.Params.Page, c.Params.PerPage)
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(jobs)
	if err != nil {
		c.Err = model.NewAppError("getJobsByType", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	w.Write(js)
}

func cancelJob(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireJobId()
	if c.Err != nil {
		return
	}

	auditRec := c.MakeAuditRecord("cancelJob", audit.Fail)
	defer c.LogAuditRec(auditRec)
	auditRec.AddEventParameter("job_id", c.Params.JobId)

	job, err := c.App.GetJob(c.Params.JobId)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.AddEventPriorState(job)
	auditRec.AddEventObjectType("job")

	// if permission to create, permission to cancel, same permission
	hasPermission, permissionRequired := c.App.SessionHasPermissionToCreateJob(*c.AppContext.Session(), job)
	if permissionRequired == nil {
		c.Err = model.NewAppError("unableToCancelJob", "api.job.unable_to_create_job.incorrect_job_type", nil, "", http.StatusBadRequest)
		return
	}

	if !hasPermission {
		c.SetPermissionError(permissionRequired)
		return
	}

	if err := c.App.CancelJob(c.Params.JobId); err != nil {
		c.Err = err
		return
	}

	auditRec.Success()

	ReturnStatusOK(w)
}
