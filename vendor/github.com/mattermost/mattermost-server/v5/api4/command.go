// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/audit"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (api *API) InitCommand() {
	api.BaseRoutes.Commands.Handle("", api.ApiSessionRequired(createCommand)).Methods("POST")
	api.BaseRoutes.Commands.Handle("", api.ApiSessionRequired(listCommands)).Methods("GET")
	api.BaseRoutes.Commands.Handle("/execute", api.ApiSessionRequired(executeCommand)).Methods("POST")

	api.BaseRoutes.Command.Handle("", api.ApiSessionRequired(getCommand)).Methods("GET")
	api.BaseRoutes.Command.Handle("", api.ApiSessionRequired(updateCommand)).Methods("PUT")
	api.BaseRoutes.Command.Handle("/move", api.ApiSessionRequired(moveCommand)).Methods("PUT")
	api.BaseRoutes.Command.Handle("", api.ApiSessionRequired(deleteCommand)).Methods("DELETE")

	api.BaseRoutes.Team.Handle("/commands/autocomplete", api.ApiSessionRequired(listAutocompleteCommands)).Methods("GET")
	api.BaseRoutes.Team.Handle("/commands/autocomplete_suggestions", api.ApiSessionRequired(listCommandAutocompleteSuggestions)).Methods("GET")
	api.BaseRoutes.Command.Handle("/regen_token", api.ApiSessionRequired(regenCommandToken)).Methods("PUT")
}

func createCommand(c *Context, w http.ResponseWriter, r *http.Request) {
	cmd := model.CommandFromJson(r.Body)
	if cmd == nil {
		c.SetInvalidParam("command")
		return
	}

	auditRec := c.MakeAuditRecord("createCommand", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), cmd.TeamId, model.PERMISSION_MANAGE_SLASH_COMMANDS) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SLASH_COMMANDS)
		return
	}

	cmd.CreatorId = c.App.Session().UserId

	rcmd, err := c.App.CreateCommand(cmd)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	c.LogAudit("success")
	auditRec.AddMeta("command", rcmd)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rcmd.ToJson()))
}

func updateCommand(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCommandId()
	if c.Err != nil {
		return
	}

	cmd := model.CommandFromJson(r.Body)
	if cmd == nil || cmd.Id != c.Params.CommandId {
		c.SetInvalidParam("command")
		return
	}

	auditRec := c.MakeAuditRecord("updateCommand", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	oldCmd, err := c.App.GetCommand(c.Params.CommandId)
	if err != nil {
		auditRec.AddMeta("command_id", c.Params.CommandId)
		c.SetCommandNotFoundError()
		return
	}
	auditRec.AddMeta("command", oldCmd)

	if cmd.TeamId != oldCmd.TeamId {
		c.Err = model.NewAppError("updateCommand", "api.command.team_mismatch.app_error", nil, "user_id="+c.App.Session().UserId, http.StatusBadRequest)
		return
	}

	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), oldCmd.TeamId, model.PERMISSION_MANAGE_SLASH_COMMANDS) {
		c.LogAudit("fail - inappropriate permissions")
		// here we return Not_found instead of a permissions error so we don't leak the existence of
		// a command to someone without permissions for the team it belongs to.
		c.SetCommandNotFoundError()
		return
	}

	if c.App.Session().UserId != oldCmd.CreatorId && !c.App.SessionHasPermissionToTeam(*c.App.Session(), oldCmd.TeamId, model.PERMISSION_MANAGE_OTHERS_SLASH_COMMANDS) {
		c.LogAudit("fail - inappropriate permissions")
		c.SetPermissionError(model.PERMISSION_MANAGE_OTHERS_SLASH_COMMANDS)
		return
	}

	rcmd, err := c.App.UpdateCommand(oldCmd, cmd)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	c.LogAudit("success")

	w.Write([]byte(rcmd.ToJson()))
}

func moveCommand(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCommandId()
	if c.Err != nil {
		return
	}

	cmr, err := model.CommandMoveRequestFromJson(r.Body)
	if err != nil {
		c.SetInvalidParam("team_id")
		return
	}

	auditRec := c.MakeAuditRecord("moveCommand", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	newTeam, appErr := c.App.GetTeam(cmr.TeamId)
	if appErr != nil {
		c.Err = appErr
		return
	}
	auditRec.AddMeta("team", newTeam)

	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), newTeam.Id, model.PERMISSION_MANAGE_SLASH_COMMANDS) {
		c.LogAudit("fail - inappropriate permissions")
		c.SetPermissionError(model.PERMISSION_MANAGE_SLASH_COMMANDS)
		return
	}

	cmd, appErr := c.App.GetCommand(c.Params.CommandId)
	if appErr != nil {
		c.SetCommandNotFoundError()
		return
	}
	auditRec.AddMeta("command", cmd)

	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), cmd.TeamId, model.PERMISSION_MANAGE_SLASH_COMMANDS) {
		c.LogAudit("fail - inappropriate permissions")
		// here we return Not_found instead of a permissions error so we don't leak the existence of
		// a command to someone without permissions for the team it belongs to.
		c.SetCommandNotFoundError()
		return
	}

	if appErr = c.App.MoveCommand(newTeam, cmd); appErr != nil {
		c.Err = appErr
		return
	}

	auditRec.Success()
	c.LogAudit("success")

	ReturnStatusOK(w)
}

func deleteCommand(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCommandId()
	if c.Err != nil {
		return
	}

	auditRec := c.MakeAuditRecord("deleteCommand", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	cmd, err := c.App.GetCommand(c.Params.CommandId)
	if err != nil {
		c.SetCommandNotFoundError()
		return
	}
	auditRec.AddMeta("command", cmd)

	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), cmd.TeamId, model.PERMISSION_MANAGE_SLASH_COMMANDS) {
		c.LogAudit("fail - inappropriate permissions")
		// here we return Not_found instead of a permissions error so we don't leak the existence of
		// a command to someone without permissions for the team it belongs to.
		c.SetCommandNotFoundError()
		return
	}

	if c.App.Session().UserId != cmd.CreatorId && !c.App.SessionHasPermissionToTeam(*c.App.Session(), cmd.TeamId, model.PERMISSION_MANAGE_OTHERS_SLASH_COMMANDS) {
		c.LogAudit("fail - inappropriate permissions")
		c.SetPermissionError(model.PERMISSION_MANAGE_OTHERS_SLASH_COMMANDS)
		return
	}

	err = c.App.DeleteCommand(cmd.Id)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	c.LogAudit("success")

	ReturnStatusOK(w)
}

func listCommands(c *Context, w http.ResponseWriter, r *http.Request) {
	customOnly, _ := strconv.ParseBool(r.URL.Query().Get("custom_only"))

	teamId := r.URL.Query().Get("team_id")
	if teamId == "" {
		c.SetInvalidParam("team_id")
		return
	}

	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), teamId, model.PERMISSION_VIEW_TEAM) {
		c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
		return
	}

	var commands []*model.Command
	var err *model.AppError
	if customOnly {
		if !c.App.SessionHasPermissionToTeam(*c.App.Session(), teamId, model.PERMISSION_MANAGE_SLASH_COMMANDS) {
			c.SetPermissionError(model.PERMISSION_MANAGE_SLASH_COMMANDS)
			return
		}
		commands, err = c.App.ListTeamCommands(teamId)
		if err != nil {
			c.Err = err
			return
		}
	} else {
		//User with no permission should see only system commands
		if !c.App.SessionHasPermissionToTeam(*c.App.Session(), teamId, model.PERMISSION_MANAGE_SLASH_COMMANDS) {
			commands, err = c.App.ListAutocompleteCommands(teamId, c.App.T)
			if err != nil {
				c.Err = err
				return
			}
		} else {
			commands, err = c.App.ListAllCommands(teamId, c.App.T)
			if err != nil {
				c.Err = err
				return
			}
		}
	}

	w.Write([]byte(model.CommandListToJson(commands)))
}

func getCommand(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCommandId()
	if c.Err != nil {
		return
	}

	cmd, err := c.App.GetCommand(c.Params.CommandId)
	if err != nil {
		c.SetCommandNotFoundError()
		return
	}

	// check for permissions to view this command; must have perms to view team and
	// PERMISSION_MANAGE_SLASH_COMMANDS for the team the command belongs to.

	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), cmd.TeamId, model.PERMISSION_VIEW_TEAM) {
		// here we return Not_found instead of a permissions error so we don't leak the existence of
		// a command to someone without permissions for the team it belongs to.
		c.SetCommandNotFoundError()
		return
	}
	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), cmd.TeamId, model.PERMISSION_MANAGE_SLASH_COMMANDS) {
		// again, return not_found to ensure id existence does not leak.
		c.SetCommandNotFoundError()
		return
	}
	w.Write([]byte(cmd.ToJson()))
}

func executeCommand(c *Context, w http.ResponseWriter, r *http.Request) {
	commandArgs := model.CommandArgsFromJson(r.Body)
	if commandArgs == nil {
		c.SetInvalidParam("command_args")
		return
	}

	if len(commandArgs.Command) <= 1 || strings.Index(commandArgs.Command, "/") != 0 || !model.IsValidId(commandArgs.ChannelId) {
		c.Err = model.NewAppError("executeCommand", "api.command.execute_command.start.app_error", nil, "", http.StatusBadRequest)
		return
	}

	auditRec := c.MakeAuditRecord("executeCommand", audit.Fail)
	defer c.LogAuditRec(auditRec)
	auditRec.AddMeta("commandargs", commandArgs)

	// checks that user is a member of the specified channel, and that they have permission to use slash commands in it
	if !c.App.SessionHasPermissionToChannel(*c.App.Session(), commandArgs.ChannelId, model.PERMISSION_USE_SLASH_COMMANDS) {
		c.SetPermissionError(model.PERMISSION_USE_SLASH_COMMANDS)
		return
	}

	channel, err := c.App.GetChannel(commandArgs.ChannelId)
	if err != nil {
		c.Err = err
		return
	}

	if channel.Type != model.CHANNEL_DIRECT && channel.Type != model.CHANNEL_GROUP {
		// if this isn't a DM or GM, the team id is implicitly taken from the channel so that slash commands created on
		// some other team can't be run against this one
		commandArgs.TeamId = channel.TeamId
	} else {
		// if the slash command was used in a DM or GM, ensure that the user is a member of the specified team, so that
		// they can't just execute slash commands against arbitrary teams
		if c.App.Session().GetTeamByTeamId(commandArgs.TeamId) == nil {
			if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_USE_SLASH_COMMANDS) {
				c.SetPermissionError(model.PERMISSION_USE_SLASH_COMMANDS)
				return
			}
		}
	}

	commandArgs.UserId = c.App.Session().UserId
	commandArgs.T = c.App.T
	commandArgs.SiteURL = c.GetSiteURLHeader()
	commandArgs.Session = *c.App.Session()

	auditRec.AddMeta("commandargs", commandArgs) // overwrite in case teamid changed

	response, err := c.App.ExecuteCommand(commandArgs)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	w.Write([]byte(response.ToJson()))
}

func listAutocompleteCommands(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTeamId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), c.Params.TeamId, model.PERMISSION_VIEW_TEAM) {
		c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
		return
	}

	commands, err := c.App.ListAutocompleteCommands(c.Params.TeamId, c.App.T)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.CommandListToJson(commands)))
}

func listCommandAutocompleteSuggestions(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTeamId()
	if c.Err != nil {
		return
	}
	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), c.Params.TeamId, model.PERMISSION_VIEW_TEAM) {
		c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
		return
	}

	roleId := model.SYSTEM_USER_ROLE_ID
	if c.IsSystemAdmin() {
		roleId = model.SYSTEM_ADMIN_ROLE_ID
	}

	query := r.URL.Query()
	userInput := query.Get("user_input")
	if userInput == "" {
		c.SetInvalidParam("userInput")
		return
	}
	userInput = strings.TrimPrefix(userInput, "/")

	commands, err := c.App.ListAutocompleteCommands(c.Params.TeamId, c.App.T)
	if err != nil {
		c.Err = err
		return
	}

	commandArgs := &model.CommandArgs{
		ChannelId: query.Get("channel_id"),
		TeamId:    c.Params.TeamId,
		RootId:    query.Get("root_id"),
		ParentId:  query.Get("parent_id"),
		UserId:    c.App.Session().UserId,
		T:         c.App.T,
		Session:   *c.App.Session(),
		SiteURL:   c.GetSiteURLHeader(),
		Command:   userInput,
	}

	suggestions := c.App.GetSuggestions(commandArgs, commands, roleId)

	w.Write(model.AutocompleteSuggestionsToJSON(suggestions))
}

func regenCommandToken(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCommandId()
	if c.Err != nil {
		return
	}

	auditRec := c.MakeAuditRecord("regenCommandToken", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	cmd, err := c.App.GetCommand(c.Params.CommandId)
	if err != nil {
		auditRec.AddMeta("command_id", c.Params.CommandId)
		c.SetCommandNotFoundError()
		return
	}
	auditRec.AddMeta("command", cmd)

	if !c.App.SessionHasPermissionToTeam(*c.App.Session(), cmd.TeamId, model.PERMISSION_MANAGE_SLASH_COMMANDS) {
		c.LogAudit("fail - inappropriate permissions")
		// here we return Not_found instead of a permissions error so we don't leak the existence of
		// a command to someone without permissions for the team it belongs to.
		c.SetCommandNotFoundError()
		return
	}

	if c.App.Session().UserId != cmd.CreatorId && !c.App.SessionHasPermissionToTeam(*c.App.Session(), cmd.TeamId, model.PERMISSION_MANAGE_OTHERS_SLASH_COMMANDS) {
		c.LogAudit("fail - inappropriate permissions")
		c.SetPermissionError(model.PERMISSION_MANAGE_OTHERS_SLASH_COMMANDS)
		return
	}

	rcmd, err := c.App.RegenCommandToken(cmd)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	c.LogAudit("success")

	resp := make(map[string]string)
	resp["token"] = rcmd.Token

	w.Write([]byte(model.MapToJson(resp)))
}
