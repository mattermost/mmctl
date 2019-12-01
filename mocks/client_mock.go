// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mmctl/client (interfaces: Client)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	model "github.com/mattermost/mattermost-server/v5/model"
	io "io"
	reflect "reflect"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// AddChannelMember mocks base method
func (m *MockClient) AddChannelMember(arg0, arg1 string) (*model.ChannelMember, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddChannelMember", arg0, arg1)
	ret0, _ := ret[0].(*model.ChannelMember)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// AddChannelMember indicates an expected call of AddChannelMember
func (mr *MockClientMockRecorder) AddChannelMember(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddChannelMember", reflect.TypeOf((*MockClient)(nil).AddChannelMember), arg0, arg1)
}

// AddTeamMember mocks base method
func (m *MockClient) AddTeamMember(arg0, arg1 string) (*model.TeamMember, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTeamMember", arg0, arg1)
	ret0, _ := ret[0].(*model.TeamMember)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// AddTeamMember indicates an expected call of AddTeamMember
func (mr *MockClientMockRecorder) AddTeamMember(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTeamMember", reflect.TypeOf((*MockClient)(nil).AddTeamMember), arg0, arg1)
}

// ConvertChannelToPrivate mocks base method
func (m *MockClient) ConvertChannelToPrivate(arg0 string) (*model.Channel, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConvertChannelToPrivate", arg0)
	ret0, _ := ret[0].(*model.Channel)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// ConvertChannelToPrivate indicates an expected call of ConvertChannelToPrivate
func (mr *MockClientMockRecorder) ConvertChannelToPrivate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConvertChannelToPrivate", reflect.TypeOf((*MockClient)(nil).ConvertChannelToPrivate), arg0)
}

// CreateChannel mocks base method
func (m *MockClient) CreateChannel(arg0 *model.Channel) (*model.Channel, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateChannel", arg0)
	ret0, _ := ret[0].(*model.Channel)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// CreateChannel indicates an expected call of CreateChannel
func (mr *MockClientMockRecorder) CreateChannel(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateChannel", reflect.TypeOf((*MockClient)(nil).CreateChannel), arg0)
}

// CreateCommand mocks base method
func (m *MockClient) CreateCommand(arg0 *model.Command) (*model.Command, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCommand", arg0)
	ret0, _ := ret[0].(*model.Command)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// CreateCommand indicates an expected call of CreateCommand
func (mr *MockClientMockRecorder) CreateCommand(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCommand", reflect.TypeOf((*MockClient)(nil).CreateCommand), arg0)
}

// CreatePost mocks base method
func (m *MockClient) CreatePost(arg0 *model.Post) (*model.Post, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePost", arg0)
	ret0, _ := ret[0].(*model.Post)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// CreatePost indicates an expected call of CreatePost
func (mr *MockClientMockRecorder) CreatePost(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePost", reflect.TypeOf((*MockClient)(nil).CreatePost), arg0)
}

// CreateTeam mocks base method
func (m *MockClient) CreateTeam(arg0 *model.Team) (*model.Team, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateTeam", arg0)
	ret0, _ := ret[0].(*model.Team)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// CreateTeam indicates an expected call of CreateTeam
func (mr *MockClientMockRecorder) CreateTeam(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTeam", reflect.TypeOf((*MockClient)(nil).CreateTeam), arg0)
}

// CreateUser mocks base method
func (m *MockClient) CreateUser(arg0 *model.User) (*model.User, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", arg0)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser
func (mr *MockClientMockRecorder) CreateUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockClient)(nil).CreateUser), arg0)
}

// DeleteChannel mocks base method
func (m *MockClient) DeleteChannel(arg0 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteChannel", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// DeleteChannel indicates an expected call of DeleteChannel
func (mr *MockClientMockRecorder) DeleteChannel(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteChannel", reflect.TypeOf((*MockClient)(nil).DeleteChannel), arg0)
}

// DeleteCommand mocks base method
func (m *MockClient) DeleteCommand(arg0 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCommand", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// DeleteCommand indicates an expected call of DeleteCommand
func (mr *MockClientMockRecorder) DeleteCommand(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCommand", reflect.TypeOf((*MockClient)(nil).DeleteCommand), arg0)
}

// DeleteUser mocks base method
func (m *MockClient) DeleteUser(arg0 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// DeleteUser indicates an expected call of DeleteUser
func (mr *MockClientMockRecorder) DeleteUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockClient)(nil).DeleteUser), arg0)
}

// DisablePlugin mocks base method
func (m *MockClient) DisablePlugin(arg0 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DisablePlugin", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// DisablePlugin indicates an expected call of DisablePlugin
func (mr *MockClientMockRecorder) DisablePlugin(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DisablePlugin", reflect.TypeOf((*MockClient)(nil).DisablePlugin), arg0)
}

// EnablePlugin mocks base method
func (m *MockClient) EnablePlugin(arg0 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnablePlugin", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// EnablePlugin indicates an expected call of EnablePlugin
func (mr *MockClientMockRecorder) EnablePlugin(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnablePlugin", reflect.TypeOf((*MockClient)(nil).EnablePlugin), arg0)
}

// GetAllTeams mocks base method
func (m *MockClient) GetAllTeams(arg0 string, arg1, arg2 int) ([]*model.Team, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllTeams", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*model.Team)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetAllTeams indicates an expected call of GetAllTeams
func (mr *MockClientMockRecorder) GetAllTeams(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllTeams", reflect.TypeOf((*MockClient)(nil).GetAllTeams), arg0, arg1, arg2)
}

// GetChannel mocks base method
func (m *MockClient) GetChannel(arg0, arg1 string) (*model.Channel, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChannel", arg0, arg1)
	ret0, _ := ret[0].(*model.Channel)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetChannel indicates an expected call of GetChannel
func (mr *MockClientMockRecorder) GetChannel(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChannel", reflect.TypeOf((*MockClient)(nil).GetChannel), arg0, arg1)
}

// GetChannelByName mocks base method
func (m *MockClient) GetChannelByName(arg0, arg1, arg2 string) (*model.Channel, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChannelByName", arg0, arg1, arg2)
	ret0, _ := ret[0].(*model.Channel)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetChannelByName indicates an expected call of GetChannelByName
func (mr *MockClientMockRecorder) GetChannelByName(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChannelByName", reflect.TypeOf((*MockClient)(nil).GetChannelByName), arg0, arg1, arg2)
}

// GetChannelByNameIncludeDeleted mocks base method
func (m *MockClient) GetChannelByNameIncludeDeleted(arg0, arg1, arg2 string) (*model.Channel, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChannelByNameIncludeDeleted", arg0, arg1, arg2)
	ret0, _ := ret[0].(*model.Channel)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetChannelByNameIncludeDeleted indicates an expected call of GetChannelByNameIncludeDeleted
func (mr *MockClientMockRecorder) GetChannelByNameIncludeDeleted(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChannelByNameIncludeDeleted", reflect.TypeOf((*MockClient)(nil).GetChannelByNameIncludeDeleted), arg0, arg1, arg2)
}

// GetChannelMembers mocks base method
func (m *MockClient) GetChannelMembers(arg0 string, arg1, arg2 int, arg3 string) (*model.ChannelMembers, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChannelMembers", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*model.ChannelMembers)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetChannelMembers indicates an expected call of GetChannelMembers
func (mr *MockClientMockRecorder) GetChannelMembers(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChannelMembers", reflect.TypeOf((*MockClient)(nil).GetChannelMembers), arg0, arg1, arg2, arg3)
}

// GetConfig mocks base method
func (m *MockClient) GetConfig() (*model.Config, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfig")
	ret0, _ := ret[0].(*model.Config)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetConfig indicates an expected call of GetConfig
func (mr *MockClientMockRecorder) GetConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfig", reflect.TypeOf((*MockClient)(nil).GetConfig))
}

// GetDeletedChannelsForTeam mocks base method
func (m *MockClient) GetDeletedChannelsForTeam(arg0 string, arg1, arg2 int, arg3 string) ([]*model.Channel, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeletedChannelsForTeam", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*model.Channel)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetDeletedChannelsForTeam indicates an expected call of GetDeletedChannelsForTeam
func (mr *MockClientMockRecorder) GetDeletedChannelsForTeam(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeletedChannelsForTeam", reflect.TypeOf((*MockClient)(nil).GetDeletedChannelsForTeam), arg0, arg1, arg2, arg3)
}

// GetGroupsByChannel mocks base method
func (m *MockClient) GetGroupsByChannel(arg0 string, arg1, arg2 int) ([]*model.Group, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroupsByChannel", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*model.Group)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetGroupsByChannel indicates an expected call of GetGroupsByChannel
func (mr *MockClientMockRecorder) GetGroupsByChannel(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupsByChannel", reflect.TypeOf((*MockClient)(nil).GetGroupsByChannel), arg0, arg1, arg2)
}

// GetGroupsByTeam mocks base method
func (m *MockClient) GetGroupsByTeam(arg0 string, arg1, arg2 int) ([]*model.Group, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroupsByTeam", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*model.Group)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetGroupsByTeam indicates an expected call of GetGroupsByTeam
func (mr *MockClientMockRecorder) GetGroupsByTeam(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupsByTeam", reflect.TypeOf((*MockClient)(nil).GetGroupsByTeam), arg0, arg1, arg2)
}

// GetLdapGroups mocks base method
func (m *MockClient) GetLdapGroups() ([]*model.Group, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLdapGroups")
	ret0, _ := ret[0].([]*model.Group)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetLdapGroups indicates an expected call of GetLdapGroups
func (mr *MockClientMockRecorder) GetLdapGroups() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLdapGroups", reflect.TypeOf((*MockClient)(nil).GetLdapGroups))
}

// GetLogs mocks base method
func (m *MockClient) GetLogs(arg0, arg1 int) ([]string, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLogs", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetLogs indicates an expected call of GetLogs
func (mr *MockClientMockRecorder) GetLogs(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLogs", reflect.TypeOf((*MockClient)(nil).GetLogs), arg0, arg1)
}

// GetPlugins mocks base method
func (m *MockClient) GetPlugins() (*model.PluginsResponse, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPlugins")
	ret0, _ := ret[0].(*model.PluginsResponse)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetPlugins indicates an expected call of GetPlugins
func (mr *MockClientMockRecorder) GetPlugins() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPlugins", reflect.TypeOf((*MockClient)(nil).GetPlugins))
}

// GetPost mocks base method
func (m *MockClient) GetPost(arg0, arg1 string) (*model.Post, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPost", arg0, arg1)
	ret0, _ := ret[0].(*model.Post)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetPost indicates an expected call of GetPost
func (mr *MockClientMockRecorder) GetPost(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPost", reflect.TypeOf((*MockClient)(nil).GetPost), arg0, arg1)
}

// GetPostsForChannel mocks base method
func (m *MockClient) GetPostsForChannel(arg0 string, arg1, arg2 int, arg3 string) (*model.PostList, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPostsForChannel", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*model.PostList)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetPostsForChannel indicates an expected call of GetPostsForChannel
func (mr *MockClientMockRecorder) GetPostsForChannel(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPostsForChannel", reflect.TypeOf((*MockClient)(nil).GetPostsForChannel), arg0, arg1, arg2, arg3)
}

// GetPublicChannelsForTeam mocks base method
func (m *MockClient) GetPublicChannelsForTeam(arg0 string, arg1, arg2 int, arg3 string) ([]*model.Channel, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPublicChannelsForTeam", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*model.Channel)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetPublicChannelsForTeam indicates an expected call of GetPublicChannelsForTeam
func (mr *MockClientMockRecorder) GetPublicChannelsForTeam(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPublicChannelsForTeam", reflect.TypeOf((*MockClient)(nil).GetPublicChannelsForTeam), arg0, arg1, arg2, arg3)
}

// GetRoleByName mocks base method
func (m *MockClient) GetRoleByName(arg0 string) (*model.Role, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRoleByName", arg0)
	ret0, _ := ret[0].(*model.Role)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetRoleByName indicates an expected call of GetRoleByName
func (mr *MockClientMockRecorder) GetRoleByName(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRoleByName", reflect.TypeOf((*MockClient)(nil).GetRoleByName), arg0)
}

// GetTeam mocks base method
func (m *MockClient) GetTeam(arg0, arg1 string) (*model.Team, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTeam", arg0, arg1)
	ret0, _ := ret[0].(*model.Team)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetTeam indicates an expected call of GetTeam
func (mr *MockClientMockRecorder) GetTeam(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTeam", reflect.TypeOf((*MockClient)(nil).GetTeam), arg0, arg1)
}

// GetTeamByName mocks base method
func (m *MockClient) GetTeamByName(arg0, arg1 string) (*model.Team, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTeamByName", arg0, arg1)
	ret0, _ := ret[0].(*model.Team)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetTeamByName indicates an expected call of GetTeamByName
func (mr *MockClientMockRecorder) GetTeamByName(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTeamByName", reflect.TypeOf((*MockClient)(nil).GetTeamByName), arg0, arg1)
}

// GetUser mocks base method
func (m *MockClient) GetUser(arg0, arg1 string) (*model.User, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", arg0, arg1)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser
func (mr *MockClientMockRecorder) GetUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockClient)(nil).GetUser), arg0, arg1)
}

// GetUserByEmail mocks base method
func (m *MockClient) GetUserByEmail(arg0, arg1 string) (*model.User, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByEmail", arg0, arg1)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetUserByEmail indicates an expected call of GetUserByEmail
func (mr *MockClientMockRecorder) GetUserByEmail(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByEmail", reflect.TypeOf((*MockClient)(nil).GetUserByEmail), arg0, arg1)
}

// GetUserByUsername mocks base method
func (m *MockClient) GetUserByUsername(arg0, arg1 string) (*model.User, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByUsername", arg0, arg1)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// GetUserByUsername indicates an expected call of GetUserByUsername
func (mr *MockClientMockRecorder) GetUserByUsername(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByUsername", reflect.TypeOf((*MockClient)(nil).GetUserByUsername), arg0, arg1)
}

// InviteUsersToTeam mocks base method
func (m *MockClient) InviteUsersToTeam(arg0 string, arg1 []string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InviteUsersToTeam", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// InviteUsersToTeam indicates an expected call of InviteUsersToTeam
func (mr *MockClientMockRecorder) InviteUsersToTeam(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InviteUsersToTeam", reflect.TypeOf((*MockClient)(nil).InviteUsersToTeam), arg0, arg1)
}

// ListCommands mocks base method
func (m *MockClient) ListCommands(arg0 string, arg1 bool) ([]*model.Command, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListCommands", arg0, arg1)
	ret0, _ := ret[0].([]*model.Command)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// ListCommands indicates an expected call of ListCommands
func (mr *MockClientMockRecorder) ListCommands(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListCommands", reflect.TypeOf((*MockClient)(nil).ListCommands), arg0, arg1)
}

// PatchChannel mocks base method
func (m *MockClient) PatchChannel(arg0 string, arg1 *model.ChannelPatch) (*model.Channel, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PatchChannel", arg0, arg1)
	ret0, _ := ret[0].(*model.Channel)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// PatchChannel indicates an expected call of PatchChannel
func (mr *MockClientMockRecorder) PatchChannel(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PatchChannel", reflect.TypeOf((*MockClient)(nil).PatchChannel), arg0, arg1)
}

// PatchRole mocks base method
func (m *MockClient) PatchRole(arg0 string, arg1 *model.RolePatch) (*model.Role, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PatchRole", arg0, arg1)
	ret0, _ := ret[0].(*model.Role)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// PatchRole indicates an expected call of PatchRole
func (mr *MockClientMockRecorder) PatchRole(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PatchRole", reflect.TypeOf((*MockClient)(nil).PatchRole), arg0, arg1)
}

// PatchTeam mocks base method
func (m *MockClient) PatchTeam(arg0 string, arg1 *model.TeamPatch) (*model.Team, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PatchTeam", arg0, arg1)
	ret0, _ := ret[0].(*model.Team)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// PatchTeam indicates an expected call of PatchTeam
func (mr *MockClientMockRecorder) PatchTeam(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PatchTeam", reflect.TypeOf((*MockClient)(nil).PatchTeam), arg0, arg1)
}

// PermanentDeleteTeam mocks base method
func (m *MockClient) PermanentDeleteTeam(arg0 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PermanentDeleteTeam", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// PermanentDeleteTeam indicates an expected call of PermanentDeleteTeam
func (mr *MockClientMockRecorder) PermanentDeleteTeam(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PermanentDeleteTeam", reflect.TypeOf((*MockClient)(nil).PermanentDeleteTeam), arg0)
}

// RemoveLicenseFile mocks base method
func (m *MockClient) RemoveLicenseFile() (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveLicenseFile")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// RemoveLicenseFile indicates an expected call of RemoveLicenseFile
func (mr *MockClientMockRecorder) RemoveLicenseFile() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveLicenseFile", reflect.TypeOf((*MockClient)(nil).RemoveLicenseFile))
}

// RemovePlugin mocks base method
func (m *MockClient) RemovePlugin(arg0 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemovePlugin", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// RemovePlugin indicates an expected call of RemovePlugin
func (mr *MockClientMockRecorder) RemovePlugin(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemovePlugin", reflect.TypeOf((*MockClient)(nil).RemovePlugin), arg0)
}

// RemoveTeamMember mocks base method
func (m *MockClient) RemoveTeamMember(arg0, arg1 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveTeamMember", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// RemoveTeamMember indicates an expected call of RemoveTeamMember
func (mr *MockClientMockRecorder) RemoveTeamMember(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveTeamMember", reflect.TypeOf((*MockClient)(nil).RemoveTeamMember), arg0, arg1)
}

// RemoveUserFromChannel mocks base method
func (m *MockClient) RemoveUserFromChannel(arg0, arg1 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveUserFromChannel", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// RemoveUserFromChannel indicates an expected call of RemoveUserFromChannel
func (mr *MockClientMockRecorder) RemoveUserFromChannel(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveUserFromChannel", reflect.TypeOf((*MockClient)(nil).RemoveUserFromChannel), arg0, arg1)
}

// RestoreChannel mocks base method
func (m *MockClient) RestoreChannel(arg0 string) (*model.Channel, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RestoreChannel", arg0)
	ret0, _ := ret[0].(*model.Channel)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// RestoreChannel indicates an expected call of RestoreChannel
func (mr *MockClientMockRecorder) RestoreChannel(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RestoreChannel", reflect.TypeOf((*MockClient)(nil).RestoreChannel), arg0)
}

// SearchTeams mocks base method
func (m *MockClient) SearchTeams(arg0 *model.TeamSearch) ([]*model.Team, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchTeams", arg0)
	ret0, _ := ret[0].([]*model.Team)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// SearchTeams indicates an expected call of SearchTeams
func (mr *MockClientMockRecorder) SearchTeams(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchTeams", reflect.TypeOf((*MockClient)(nil).SearchTeams), arg0)
}

// SendPasswordResetEmail mocks base method
func (m *MockClient) SendPasswordResetEmail(arg0 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendPasswordResetEmail", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// SendPasswordResetEmail indicates an expected call of SendPasswordResetEmail
func (mr *MockClientMockRecorder) SendPasswordResetEmail(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendPasswordResetEmail", reflect.TypeOf((*MockClient)(nil).SendPasswordResetEmail), arg0)
}

// SyncLdap mocks base method
func (m *MockClient) SyncLdap() (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncLdap")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// SyncLdap indicates an expected call of SyncLdap
func (mr *MockClientMockRecorder) SyncLdap() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncLdap", reflect.TypeOf((*MockClient)(nil).SyncLdap))
}

// UpdateUser mocks base method
func (m *MockClient) UpdateUser(arg0 *model.User) (*model.User, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", arg0)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// UpdateUser indicates an expected call of UpdateUser
func (mr *MockClientMockRecorder) UpdateUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockClient)(nil).UpdateUser), arg0)
}

// UpdateUserMfa mocks base method
func (m *MockClient) UpdateUserMfa(arg0, arg1 string, arg2 bool) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserMfa", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// UpdateUserMfa indicates an expected call of UpdateUserMfa
func (mr *MockClientMockRecorder) UpdateUserMfa(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserMfa", reflect.TypeOf((*MockClient)(nil).UpdateUserMfa), arg0, arg1, arg2)
}

// UpdateUserRoles mocks base method
func (m *MockClient) UpdateUserRoles(arg0, arg1 string) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserRoles", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// UpdateUserRoles indicates an expected call of UpdateUserRoles
func (mr *MockClientMockRecorder) UpdateUserRoles(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserRoles", reflect.TypeOf((*MockClient)(nil).UpdateUserRoles), arg0, arg1)
}

// UploadLicenseFile mocks base method
func (m *MockClient) UploadLicenseFile(arg0 []byte) (bool, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadLicenseFile", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// UploadLicenseFile indicates an expected call of UploadLicenseFile
func (mr *MockClientMockRecorder) UploadLicenseFile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadLicenseFile", reflect.TypeOf((*MockClient)(nil).UploadLicenseFile), arg0)
}

// UploadPlugin mocks base method
func (m *MockClient) UploadPlugin(arg0 io.Reader) (*model.Manifest, *model.Response) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadPlugin", arg0)
	ret0, _ := ret[0].(*model.Manifest)
	ret1, _ := ret[1].(*model.Response)
	return ret0, ret1
}

// UploadPlugin indicates an expected call of UploadPlugin
func (mr *MockClientMockRecorder) UploadPlugin(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadPlugin", reflect.TypeOf((*MockClient)(nil).UploadPlugin), arg0)
}
