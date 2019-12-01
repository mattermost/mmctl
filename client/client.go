package client

import (
	"io"

	"github.com/mattermost/mattermost-server/v5/model"
)

type Client interface {
	CreateChannel(channel *model.Channel) (*model.Channel, *model.Response)
	RemoveUserFromChannel(channelId, userId string) (bool, *model.Response)
	GetChannelMembers(channelId string, page, perPage int, etag string) (*model.ChannelMembers, *model.Response)
	AddChannelMember(channelId, userId string) (*model.ChannelMember, *model.Response)
	DeleteChannel(channelId string) (bool, *model.Response)
	GetPublicChannelsForTeam(teamId string, page int, perPage int, etag string) ([]*model.Channel, *model.Response)
	GetDeletedChannelsForTeam(teamId string, page int, perPage int, etag string) ([]*model.Channel, *model.Response)
	RestoreChannel(channelId string) (*model.Channel, *model.Response)
	ConvertChannelToPrivate(channelId string) (*model.Channel, *model.Response)
	PatchChannel(channelId string, patch *model.ChannelPatch) (*model.Channel, *model.Response)
	GetChannelByName(channelName, teamId string, etag string) (*model.Channel, *model.Response)
	GetChannelByNameIncludeDeleted(channelName, teamId string, etag string) (*model.Channel, *model.Response)
	GetChannel(channelId, etag string) (*model.Channel, *model.Response)
	GetTeam(teamId, etag string) (*model.Team, *model.Response)
	GetTeamByName(name, etag string) (*model.Team, *model.Response)
	GetAllTeams(etag string, page int, perPage int) ([]*model.Team, *model.Response)
	CreateTeam(team *model.Team) (*model.Team, *model.Response)
	PatchTeam(teamId string, patch *model.TeamPatch) (*model.Team, *model.Response)
	AddTeamMember(teamId, userId string) (*model.TeamMember, *model.Response)
	RemoveTeamMember(teamId, userId string) (bool, *model.Response)
	PermanentDeleteTeam(teamId string) (bool, *model.Response)
	SearchTeams(search *model.TeamSearch) ([]*model.Team, *model.Response)
	GetPost(postId string, etag string) (*model.Post, *model.Response)
	CreatePost(post *model.Post) (*model.Post, *model.Response)
	GetPostsForChannel(channelId string, page, perPage int, etag string) (*model.PostList, *model.Response)
	GetLdapGroups() ([]*model.Group, *model.Response)
	GetGroupsByChannel(channelId string, page, perPage int) ([]*model.Group, *model.Response)
	GetGroupsByTeam(teamId string, page, perPage int) ([]*model.Group, *model.Response)
	UploadLicenseFile(data []byte) (bool, *model.Response)
	RemoveLicenseFile() (bool, *model.Response)
	GetLogs(page, perPage int) ([]string, *model.Response)
	GetRoleByName(name string) (*model.Role, *model.Response)
	PatchRole(roleId string, patch *model.RolePatch) (*model.Role, *model.Response)
	UploadPlugin(file io.Reader) (*model.Manifest, *model.Response)
	RemovePlugin(id string) (bool, *model.Response)
	EnablePlugin(id string) (bool, *model.Response)
	DisablePlugin(id string) (bool, *model.Response)
	GetPlugins() (*model.PluginsResponse, *model.Response)
	GetUser(userId, etag string) (*model.User, *model.Response)
	GetUserByUsername(userName, etag string) (*model.User, *model.Response)
	GetUserByEmail(email, etag string) (*model.User, *model.Response)
	CreateUser(user *model.User) (*model.User, *model.Response)
	DeleteUser(userId string) (bool, *model.Response)
	UpdateUserRoles(userId, roles string) (bool, *model.Response)
	InviteUsersToTeam(teamId string, userEmails []string) (bool, *model.Response)
	SendPasswordResetEmail(email string) (bool, *model.Response)
	UpdateUser(user *model.User) (*model.User, *model.Response)
	UpdateUserMfa(userId, code string, activate bool) (bool, *model.Response)
	CreateCommand(cmd *model.Command) (*model.Command, *model.Response)
	ListCommands(teamId string, customOnly bool) ([]*model.Command, *model.Response)
	DeleteCommand(commandId string) (bool, *model.Response)
	GetConfig() (*model.Config, *model.Response)
	SyncLdap() (bool, *model.Response)
}
