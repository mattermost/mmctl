// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/mattermost/mattermost-server/v6/app"
	"github.com/mattermost/mattermost-server/v6/audit"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

func (api *API) InitPost() {
	api.BaseRoutes.Posts.Handle("", api.APISessionRequired(createPost)).Methods("POST")
	api.BaseRoutes.Post.Handle("", api.APISessionRequired(getPost)).Methods("GET")
	api.BaseRoutes.Post.Handle("", api.APISessionRequired(deletePost)).Methods("DELETE")
	api.BaseRoutes.Posts.Handle("/ids", api.APISessionRequired(getPostsByIds)).Methods("POST")
	api.BaseRoutes.Posts.Handle("/ephemeral", api.APISessionRequired(createEphemeralPost)).Methods("POST")
	api.BaseRoutes.Post.Handle("/thread", api.APISessionRequired(getPostThread)).Methods("GET")
	api.BaseRoutes.Post.Handle("/files/info", api.APISessionRequired(getFileInfosForPost)).Methods("GET")
	api.BaseRoutes.PostsForChannel.Handle("", api.APISessionRequired(getPostsForChannel)).Methods("GET")
	api.BaseRoutes.PostsForUser.Handle("/flagged", api.APISessionRequired(getFlaggedPostsForUser)).Methods("GET")

	api.BaseRoutes.ChannelForUser.Handle("/posts/unread", api.APISessionRequired(getPostsForChannelAroundLastUnread)).Methods("GET")

	api.BaseRoutes.Team.Handle("/posts/search", api.APISessionRequiredDisableWhenBusy(searchPostsInTeam)).Methods("POST")
	api.BaseRoutes.Posts.Handle("/search", api.APISessionRequiredDisableWhenBusy(searchPostsInAllTeams)).Methods("POST")
	api.BaseRoutes.Post.Handle("", api.APISessionRequired(updatePost)).Methods("PUT")
	api.BaseRoutes.Post.Handle("/patch", api.APISessionRequired(patchPost)).Methods("PUT")
	api.BaseRoutes.PostForUser.Handle("/set_unread", api.APISessionRequired(setPostUnread)).Methods("POST")
	api.BaseRoutes.Post.Handle("/pin", api.APISessionRequired(pinPost)).Methods("POST")
	api.BaseRoutes.Post.Handle("/unpin", api.APISessionRequired(unpinPost)).Methods("POST")
}

func createPost(c *Context, w http.ResponseWriter, r *http.Request) {
	var post model.Post
	if jsonErr := json.NewDecoder(r.Body).Decode(&post); jsonErr != nil {
		c.SetInvalidParam("post")
		return
	}

	// Strip away delete_at if passed
	post.DeleteAt = 0

	post.UserId = c.AppContext.Session().UserId

	auditRec := c.MakeAuditRecord("createPost", audit.Fail)
	defer c.LogAuditRecWithLevel(auditRec, app.LevelContent)
	auditRec.AddMeta("post", &post)

	hasPermission := false
	if c.App.SessionHasPermissionToChannel(*c.AppContext.Session(), post.ChannelId, model.PermissionCreatePost) {
		hasPermission = true
	} else if channel, err := c.App.GetChannel(post.ChannelId); err == nil {
		// Temporary permission check method until advanced permissions, please do not copy
		if channel.Type == model.ChannelTypeOpen && c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), channel.TeamId, model.PermissionCreatePostPublic) {
			hasPermission = true
		}
	}

	if !hasPermission {
		c.SetPermissionError(model.PermissionCreatePost)
		return
	}

	if post.CreateAt != 0 && !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionManageSystem) {
		post.CreateAt = 0
	}

	setOnline := r.URL.Query().Get("set_online")
	setOnlineBool := true // By default, always set online.
	var err2 error
	if setOnline != "" {
		setOnlineBool, err2 = strconv.ParseBool(setOnline)
		if err2 != nil {
			mlog.Warn("Failed to parse set_online URL query parameter from createPost request", mlog.Err(err2))
			setOnlineBool = true // Set online nevertheless.
		}
	}

	rp, err := c.App.CreatePostAsUser(c.AppContext, c.App.PostWithProxyRemovedFromImageURLs(&post), c.AppContext.Session().Id, setOnlineBool)
	if err != nil {
		c.Err = err
		return
	}
	auditRec.Success()
	auditRec.AddMeta("post", rp) // overwrite meta

	if setOnlineBool {
		c.App.SetStatusOnline(c.AppContext.Session().UserId, false)
	}

	c.App.UpdateLastActivityAtIfNeeded(*c.AppContext.Session())
	c.ExtendSessionExpiryIfNeeded(w, r)

	w.WriteHeader(http.StatusCreated)

	// Note that rp has already had PreparePostForClient called on it by App.CreatePost
	if err := json.NewEncoder(w).Encode(rp); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func createEphemeralPost(c *Context, w http.ResponseWriter, r *http.Request) {
	ephRequest := model.PostEphemeral{}

	json.NewDecoder(r.Body).Decode(&ephRequest)
	if ephRequest.UserID == "" {
		c.SetInvalidParam("user_id")
		return
	}

	if ephRequest.Post == nil {
		c.SetInvalidParam("post")
		return
	}

	ephRequest.Post.UserId = c.AppContext.Session().UserId
	ephRequest.Post.CreateAt = model.GetMillis()

	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionCreatePostEphemeral) {
		c.SetPermissionError(model.PermissionCreatePostEphemeral)
		return
	}

	rp := c.App.SendEphemeralPost(ephRequest.UserID, c.App.PostWithProxyRemovedFromImageURLs(ephRequest.Post))

	w.WriteHeader(http.StatusCreated)
	rp = model.AddPostActionCookies(rp, c.App.PostActionCookieSecret())
	rp = c.App.PreparePostForClientWithEmbedsAndImages(rp, true, false)
	rp, err := c.App.SanitizePostMetadataForUser(rp, c.AppContext.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}
	if err := json.NewEncoder(w).Encode(rp); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func getPostsForChannel(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireChannelId()
	if c.Err != nil {
		return
	}

	afterPost := r.URL.Query().Get("after")
	if afterPost != "" && !model.IsValidId(afterPost) {
		c.SetInvalidParam("after")
		return
	}

	beforePost := r.URL.Query().Get("before")
	if beforePost != "" && !model.IsValidId(beforePost) {
		c.SetInvalidParam("before")
		return
	}

	sinceString := r.URL.Query().Get("since")
	var since int64
	var parseError error
	if sinceString != "" {
		since, parseError = strconv.ParseInt(sinceString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("since")
			return
		}
	}
	skipFetchThreads := r.URL.Query().Get("skipFetchThreads") == "true"
	collapsedThreads := r.URL.Query().Get("collapsedThreads") == "true"
	collapsedThreadsExtended := r.URL.Query().Get("collapsedThreadsExtended") == "true"
	channelId := c.Params.ChannelId
	page := c.Params.Page
	perPage := c.Params.PerPage

	if !c.App.SessionHasPermissionToChannel(*c.AppContext.Session(), channelId, model.PermissionReadChannel) {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	if !*c.App.Config().TeamSettings.ExperimentalViewArchivedChannels {
		channel, err := c.App.GetChannel(channelId)
		if err != nil {
			c.Err = err
			return
		}
		if channel.DeleteAt != 0 {
			c.Err = model.NewAppError("Api4.getPostsForChannel", "api.user.view_archived_channels.get_posts_for_channel.app_error", nil, "", http.StatusForbidden)
			return
		}
	}

	var list *model.PostList
	var err *model.AppError
	etag := ""

	if since > 0 {
		list, err = c.App.GetPostsSince(model.GetPostsSinceOptions{ChannelId: channelId, Time: since, SkipFetchThreads: skipFetchThreads, CollapsedThreads: collapsedThreads, CollapsedThreadsExtended: collapsedThreadsExtended, UserId: c.AppContext.Session().UserId})
	} else if afterPost != "" {
		etag = c.App.GetPostsEtag(channelId, collapsedThreads)

		if c.HandleEtag(etag, "Get Posts After", w, r) {
			return
		}

		list, err = c.App.GetPostsAfterPost(model.GetPostsOptions{ChannelId: channelId, PostId: afterPost, Page: page, PerPage: perPage, SkipFetchThreads: skipFetchThreads, CollapsedThreads: collapsedThreads, UserId: c.AppContext.Session().UserId})
	} else if beforePost != "" {
		etag = c.App.GetPostsEtag(channelId, collapsedThreads)

		if c.HandleEtag(etag, "Get Posts Before", w, r) {
			return
		}

		list, err = c.App.GetPostsBeforePost(model.GetPostsOptions{ChannelId: channelId, PostId: beforePost, Page: page, PerPage: perPage, SkipFetchThreads: skipFetchThreads, CollapsedThreads: collapsedThreads, CollapsedThreadsExtended: collapsedThreadsExtended, UserId: c.AppContext.Session().UserId})
	} else {
		etag = c.App.GetPostsEtag(channelId, collapsedThreads)

		if c.HandleEtag(etag, "Get Posts", w, r) {
			return
		}

		list, err = c.App.GetPostsPage(model.GetPostsOptions{ChannelId: channelId, Page: page, PerPage: perPage, SkipFetchThreads: skipFetchThreads, CollapsedThreads: collapsedThreads, CollapsedThreadsExtended: collapsedThreadsExtended, UserId: c.AppContext.Session().UserId})
	}

	if err != nil {
		c.Err = err
		return
	}

	if etag != "" {
		w.Header().Set(model.HeaderEtagServer, etag)
	}

	c.App.AddCursorIdsForPostList(list, afterPost, beforePost, since, page, perPage, collapsedThreads)
	clientPostList := c.App.PreparePostListForClient(list)
	clientPostList, err = c.App.SanitizePostListMetadataForUser(clientPostList, c.AppContext.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}

	if err := json.NewEncoder(w).Encode(clientPostList); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func getPostsForChannelAroundLastUnread(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId().RequireChannelId()
	if c.Err != nil {
		return
	}

	userId := c.Params.UserId
	if !c.App.SessionHasPermissionToUser(*c.AppContext.Session(), userId) {
		c.SetPermissionError(model.PermissionEditOtherUsers)
		return
	}

	channelId := c.Params.ChannelId
	if !c.App.SessionHasPermissionToChannel(*c.AppContext.Session(), channelId, model.PermissionReadChannel) {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	if c.Params.LimitAfter == 0 {
		c.SetInvalidURLParam("limit_after")
		return
	}

	skipFetchThreads := r.URL.Query().Get("skipFetchThreads") == "true"
	collapsedThreads := r.URL.Query().Get("collapsedThreads") == "true"
	collapsedThreadsExtended := r.URL.Query().Get("collapsedThreadsExtended") == "true"

	postList, err := c.App.GetPostsForChannelAroundLastUnread(channelId, userId, c.Params.LimitBefore, c.Params.LimitAfter, skipFetchThreads, collapsedThreads, collapsedThreadsExtended)
	if err != nil {
		c.Err = err
		return
	}

	etag := ""
	if len(postList.Order) == 0 {
		etag = c.App.GetPostsEtag(channelId, collapsedThreads)

		if c.HandleEtag(etag, "Get Posts", w, r) {
			return
		}

		postList, err = c.App.GetPostsPage(model.GetPostsOptions{ChannelId: channelId, Page: app.PageDefault, PerPage: c.Params.LimitBefore, SkipFetchThreads: skipFetchThreads, CollapsedThreads: collapsedThreads, CollapsedThreadsExtended: collapsedThreadsExtended, UserId: c.AppContext.Session().UserId})
		if err != nil {
			c.Err = err
			return
		}
	}

	postList.NextPostId = c.App.GetNextPostIdFromPostList(postList, collapsedThreads)
	postList.PrevPostId = c.App.GetPrevPostIdFromPostList(postList, collapsedThreads)

	clientPostList := c.App.PreparePostListForClient(postList)
	clientPostList, err = c.App.SanitizePostListMetadataForUser(clientPostList, c.AppContext.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}

	if etag != "" {
		w.Header().Set(model.HeaderEtagServer, etag)
	}
	if err := json.NewEncoder(w).Encode(clientPostList); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func getFlaggedPostsForUser(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToUser(*c.AppContext.Session(), c.Params.UserId) {
		c.SetPermissionError(model.PermissionEditOtherUsers)
		return
	}

	channelId := r.URL.Query().Get("channel_id")
	teamId := r.URL.Query().Get("team_id")

	var posts *model.PostList
	var err *model.AppError

	if channelId != "" {
		posts, err = c.App.GetFlaggedPostsForChannel(c.Params.UserId, channelId, c.Params.Page, c.Params.PerPage)
	} else if teamId != "" {
		posts, err = c.App.GetFlaggedPostsForTeam(c.Params.UserId, teamId, c.Params.Page, c.Params.PerPage)
	} else {
		posts, err = c.App.GetFlaggedPosts(c.Params.UserId, c.Params.Page, c.Params.PerPage)
	}
	if err != nil {
		c.Err = err
		return
	}

	pl := model.NewPostList()
	channelReadPermission := make(map[string]bool)

	for _, post := range posts.Posts {
		allowed, ok := channelReadPermission[post.ChannelId]

		if !ok {
			allowed = false

			if c.App.SessionHasPermissionToChannel(*c.AppContext.Session(), post.ChannelId, model.PermissionReadChannel) {
				allowed = true
			}

			channelReadPermission[post.ChannelId] = allowed
		}

		if !allowed {
			continue
		}

		pl.AddPost(post)
		pl.AddOrder(post.Id)
	}

	pl.SortByCreateAt()
	clientPostList := c.App.PreparePostListForClient(pl)
	clientPostList, err = c.App.SanitizePostListMetadataForUser(clientPostList, c.AppContext.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}
	if err := json.NewEncoder(w).Encode(clientPostList); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func getPost(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}

	post, err := c.App.GetPostIfAuthorized(c.Params.PostId, c.AppContext.Session())
	if err != nil {
		c.Err = err
		return
	}

	post = c.App.PreparePostForClientWithEmbedsAndImages(post, false, false)
	post, err = c.App.SanitizePostMetadataForUser(post, c.AppContext.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}

	if c.HandleEtag(post.Etag(), "Get Post", w, r) {
		return
	}

	w.Header().Set(model.HeaderEtagServer, post.Etag())
	if err := json.NewEncoder(w).Encode(post); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func getPostsByIds(c *Context, w http.ResponseWriter, r *http.Request) {
	postIDs := model.ArrayFromJSON(r.Body)

	if len(postIDs) == 0 {
		c.SetInvalidParam("post_ids")
		return
	}

	if len(postIDs) > 1000 {
		c.Err = model.NewAppError("getPostsByIds", "api.post.posts_by_ids.invalid_body.request_error", map[string]interface{}{"MaxLength": 1000}, "", http.StatusBadRequest)
		return
	}

	postsList, err := c.App.GetPostsByIds(postIDs)
	if err != nil {
		c.Err = err
		return
	}

	var posts = []*model.Post{}
	channelMap := make(map[string]*model.Channel)

	for _, post := range postsList {
		var channel *model.Channel
		if val, ok := channelMap[post.ChannelId]; ok {
			channel = val
		} else {
			channel, err = c.App.GetChannel(post.ChannelId)
			if err != nil {
				c.Err = err
				return
			}
			channelMap[channel.Id] = channel
		}

		if !c.App.SessionHasPermissionToChannel(*c.AppContext.Session(), channel.Id, model.PermissionReadChannel) {
			if channel.Type != model.ChannelTypeOpen || (channel.Type == model.ChannelTypeOpen && !c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), channel.TeamId, model.PermissionReadPublicChannel)) {
				continue
			}
		}

		post = c.App.PreparePostForClient(post, false, false)

		posts = append(posts, post)
	}

	if err := json.NewEncoder(w).Encode(posts); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func deletePost(c *Context, w http.ResponseWriter, _ *http.Request) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}

	auditRec := c.MakeAuditRecord("deletePost", audit.Fail)
	defer c.LogAuditRecWithLevel(auditRec, app.LevelContent)
	auditRec.AddMeta("post_id", c.Params.PostId)

	post, err := c.App.GetSinglePost(c.Params.PostId)
	if err != nil {
		c.SetPermissionError(model.PermissionDeletePost)
		return
	}
	auditRec.AddMeta("post", post)

	if c.AppContext.Session().UserId == post.UserId {
		if !c.App.SessionHasPermissionToChannel(*c.AppContext.Session(), post.ChannelId, model.PermissionDeletePost) {
			c.SetPermissionError(model.PermissionDeletePost)
			return
		}
	} else {
		if !c.App.SessionHasPermissionToChannel(*c.AppContext.Session(), post.ChannelId, model.PermissionDeleteOthersPosts) {
			c.SetPermissionError(model.PermissionDeleteOthersPosts)
			return
		}
	}

	if _, err := c.App.DeletePost(c.Params.PostId, c.AppContext.Session().UserId); err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	ReturnStatusOK(w)
}

func getPostThread(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}
	skipFetchThreads := r.URL.Query().Get("skipFetchThreads") == "true"
	collapsedThreads := r.URL.Query().Get("collapsedThreads") == "true"
	collapsedThreadsExtended := r.URL.Query().Get("collapsedThreadsExtended") == "true"
	list, err := c.App.GetPostThread(c.Params.PostId, skipFetchThreads, collapsedThreads, collapsedThreadsExtended, c.AppContext.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}

	post, ok := list.Posts[c.Params.PostId]
	if !ok {
		c.SetInvalidURLParam("post_id")
		return
	}

	if _, err = c.App.GetPostIfAuthorized(post.Id, c.AppContext.Session()); err != nil {
		c.Err = err
		return
	}

	if c.HandleEtag(list.Etag(), "Get Post Thread", w, r) {
		return
	}

	clientPostList := c.App.PreparePostListForClient(list)
	clientPostList, err = c.App.SanitizePostListMetadataForUser(clientPostList, c.AppContext.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}

	w.Header().Set(model.HeaderEtagServer, clientPostList.Etag())

	if err := json.NewEncoder(w).Encode(clientPostList); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func searchPostsInTeam(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTeamId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), c.Params.TeamId, model.PermissionViewTeam) {
		c.SetPermissionError(model.PermissionViewTeam)
		return
	}

	searchPosts(c, w, r, c.Params.TeamId)
}

func searchPostsInAllTeams(c *Context, w http.ResponseWriter, r *http.Request) {
	searchPosts(c, w, r, "")
}

func searchPosts(c *Context, w http.ResponseWriter, r *http.Request, teamId string) {
	var params model.SearchParameter
	if jsonErr := json.NewDecoder(r.Body).Decode(&params); jsonErr != nil {
		c.Err = model.NewAppError("searchPosts", "api.post.search_posts.invalid_body.app_error", nil, jsonErr.Error(), http.StatusBadRequest)
		return
	}

	if params.Terms == nil || *params.Terms == "" {
		c.SetInvalidParam("terms")
		return
	}
	terms := *params.Terms

	timeZoneOffset := 0
	if params.TimeZoneOffset != nil {
		timeZoneOffset = *params.TimeZoneOffset
	}

	isOrSearch := false
	if params.IsOrSearch != nil {
		isOrSearch = *params.IsOrSearch
	}

	page := 0
	if params.Page != nil {
		page = *params.Page
	}

	perPage := 60
	if params.PerPage != nil {
		perPage = *params.PerPage
	}

	includeDeletedChannels := false
	if params.IncludeDeletedChannels != nil {
		includeDeletedChannels = *params.IncludeDeletedChannels
	}

	startTime := time.Now()

	results, err := c.App.SearchPostsForUser(c.AppContext, terms, c.AppContext.Session().UserId, teamId, isOrSearch, includeDeletedChannels, timeZoneOffset, page, perPage)

	elapsedTime := float64(time.Since(startTime)) / float64(time.Second)
	metrics := c.App.Metrics()
	if metrics != nil {
		metrics.IncrementPostsSearchCounter()
		metrics.ObservePostsSearchDuration(elapsedTime)
	}

	if err != nil {
		c.Err = err
		return
	}

	clientPostList := c.App.PreparePostListForClient(results.PostList)
	clientPostList, err = c.App.SanitizePostListMetadataForUser(clientPostList, c.AppContext.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}

	results = model.MakePostSearchResults(clientPostList, results.Matches)

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func updatePost(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}

	var post model.Post
	if jsonErr := json.NewDecoder(r.Body).Decode(&post); jsonErr != nil {
		c.SetInvalidParam("post")
		return
	}

	auditRec := c.MakeAuditRecord("updatePost", audit.Fail)
	defer c.LogAuditRecWithLevel(auditRec, app.LevelContent)

	// The post being updated in the payload must be the same one as indicated in the URL.
	if post.Id != c.Params.PostId {
		c.SetInvalidParam("id")
		return
	}

	if !c.App.SessionHasPermissionToChannelByPost(*c.AppContext.Session(), c.Params.PostId, model.PermissionEditPost) {
		c.SetPermissionError(model.PermissionEditPost)
		return
	}

	originalPost, err := c.App.GetSinglePost(c.Params.PostId)
	if err != nil {
		c.SetPermissionError(model.PermissionEditPost)
		return
	}
	auditRec.AddMeta("post", originalPost)

	// Updating the file_ids of a post is not a supported operation and will be ignored
	post.FileIds = originalPost.FileIds

	if c.AppContext.Session().UserId != originalPost.UserId {
		if !c.App.SessionHasPermissionToChannelByPost(*c.AppContext.Session(), c.Params.PostId, model.PermissionEditOthersPosts) {
			c.SetPermissionError(model.PermissionEditOthersPosts)
			return
		}
	}

	post.Id = c.Params.PostId

	rpost, err := c.App.UpdatePost(c.AppContext, c.App.PostWithProxyRemovedFromImageURLs(&post), false)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	auditRec.AddMeta("update", rpost)

	if err := json.NewEncoder(w).Encode(rpost); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func patchPost(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}

	var post model.PostPatch
	if jsonErr := json.NewDecoder(r.Body).Decode(&post); jsonErr != nil {
		c.SetInvalidParam("post")
		return
	}

	auditRec := c.MakeAuditRecord("patchPost", audit.Fail)
	defer c.LogAuditRecWithLevel(auditRec, app.LevelContent)

	// Updating the file_ids of a post is not a supported operation and will be ignored
	post.FileIds = nil

	originalPost, err := c.App.GetSinglePost(c.Params.PostId)
	if err != nil {
		c.SetPermissionError(model.PermissionEditPost)
		return
	}
	auditRec.AddMeta("post", originalPost)

	var permission *model.Permission
	if c.AppContext.Session().UserId == originalPost.UserId {
		permission = model.PermissionEditPost
	} else {
		permission = model.PermissionEditOthersPosts
	}

	if !c.App.SessionHasPermissionToChannelByPost(*c.AppContext.Session(), c.Params.PostId, permission) {
		c.SetPermissionError(permission)
		return
	}

	patchedPost, err := c.App.PatchPost(c.AppContext, c.Params.PostId, c.App.PostPatchWithProxyRemovedFromImageURLs(&post))
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	auditRec.AddMeta("patch", patchedPost)

	if err := json.NewEncoder(w).Encode(patchedPost); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func setPostUnread(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId().RequireUserId()
	if c.Err != nil {
		return
	}

	props := model.MapBoolFromJSON(r.Body)
	collapsedThreadsSupported := props["collapsed_threads_supported"]

	if c.AppContext.Session().UserId != c.Params.UserId && !c.App.SessionHasPermissionToUser(*c.AppContext.Session(), c.Params.UserId) {
		c.SetPermissionError(model.PermissionEditOtherUsers)
		return
	}
	if !c.App.SessionHasPermissionToChannelByPost(*c.AppContext.Session(), c.Params.PostId, model.PermissionReadChannel) {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	state, err := c.App.MarkChannelAsUnreadFromPost(c.Params.PostId, c.Params.UserId, collapsedThreadsSupported, false)
	if err != nil {
		c.Err = err
		return
	}
	if err := json.NewEncoder(w).Encode(state); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}

func saveIsPinnedPost(c *Context, w http.ResponseWriter, isPinned bool) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}

	auditRec := c.MakeAuditRecord("saveIsPinnedPost", audit.Fail)
	defer c.LogAuditRecWithLevel(auditRec, app.LevelContent)

	if !c.App.SessionHasPermissionToChannelByPost(*c.AppContext.Session(), c.Params.PostId, model.PermissionReadChannel) {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	post, err := c.App.GetSinglePost(c.Params.PostId)
	if err != nil {
		c.Err = err
		return
	}
	auditRec.AddMeta("post", post)

	patch := &model.PostPatch{}
	patch.IsPinned = model.NewBool(isPinned)

	patchedPost, err := c.App.PatchPost(c.AppContext, c.Params.PostId, patch)
	if err != nil {
		c.Err = err
		return
	}
	auditRec.AddMeta("patch", patchedPost)

	auditRec.Success()
	ReturnStatusOK(w)
}

func pinPost(c *Context, w http.ResponseWriter, _ *http.Request) {
	saveIsPinnedPost(c, w, true)
}

func unpinPost(c *Context, w http.ResponseWriter, _ *http.Request) {
	saveIsPinnedPost(c, w, false)
}

func getFileInfosForPost(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToChannelByPost(*c.AppContext.Session(), c.Params.PostId, model.PermissionReadChannel) {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	infos, err := c.App.GetFileInfosForPostWithMigration(c.Params.PostId)
	if err != nil {
		c.Err = err
		return
	}

	if c.HandleEtag(model.GetEtagForFileInfos(infos), "Get File Infos For Post", w, r) {
		return
	}

	js, jsonErr := json.Marshal(infos)
	if jsonErr != nil {
		c.Err = model.NewAppError("getFileInfosForPost", "api.marshal_error", nil, jsonErr.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "max-age=2592000, private")
	w.Header().Set(model.HeaderEtagServer, model.GetEtagForFileInfos(infos))
	w.Write(js)
}
