// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/mailru/easygo/netpoll"
	goi18n "github.com/mattermost/go-i18n/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	sendQueueSize          = 256
	sendSlowWarn           = (sendQueueSize * 50) / 100
	sendFullWarn           = (sendQueueSize * 95) / 100
	writeWaitTime          = 30 * time.Second
	pongWaitTime           = 100 * time.Second
	pingInterval           = (pongWaitTime * 6) / 10
	authCheckInterval      = 5 * time.Second
	webConnMemberCacheTime = 1000 * 60 * 30 // 30 minutes
)

// WebConn represents a single websocket connection to a user.
// It contains all the necessary state to manage sending/receiving data to/from
// a websocket.
type WebConn struct {
	sessionExpiresAt int64 // This should stay at the top for 64-bit alignment of 64-bit words accessed atomically
	App              *App
	WebSocket        net.Conn
	T                goi18n.TranslateFunc
	Locale           string
	Sequence         int64
	UserId           string

	readMut                   sync.Mutex
	allChannelMembers         map[string]string
	lastAllChannelMembersTime int64
	lastUserActivityAt        int64
	send                      chan model.WebSocketMessage
	sessionToken              atomic.Value
	session                   atomic.Value
	isWindows                 bool
	endWritePump              chan struct{}
	pumpFinished              chan struct{}
}

// NewWebConn returns a new WebConn instance.
func (a *App) NewWebConn(ws net.Conn, session model.Session, t goi18n.TranslateFunc, locale string) *WebConn {
	if session.UserId != "" {
		a.Srv().Go(func() {
			a.SetStatusOnline(session.UserId, false)
			a.UpdateLastActivityAtIfNeeded(session)
		})
	}

	wc := &WebConn{
		App:                a,
		send:               make(chan model.WebSocketMessage, sendQueueSize),
		WebSocket:          ws,
		lastUserActivityAt: model.GetMillis(),
		UserId:             session.UserId,
		T:                  t,
		Locale:             locale,
		isWindows:          runtime.GOOS == "windows",
		endWritePump:       make(chan struct{}),
		pumpFinished:       make(chan struct{}),
	}

	wc.SetSession(&session)
	wc.SetSessionToken(session.Token)
	wc.SetSessionExpiresAt(session.ExpiresAt)

	// epoll/kqueue is not available on Windows.
	if !wc.isWindows {
		wc.startPoller()
	}
	return wc
}

// Close closes the WebConn.
func (wc *WebConn) Close() {
	wc.WebSocket.Close()
	if !wc.isWindows {
		// This triggers the pump exit.
		// If the pump has already exited, this just becomes a noop.
		close(wc.endWritePump)
	}
	// We wait for the pump to fully exit.
	<-wc.pumpFinished
}

// GetSessionExpiresAt returns the time at which the session expires.
func (wc *WebConn) GetSessionExpiresAt() int64 {
	return atomic.LoadInt64(&wc.sessionExpiresAt)
}

// SetSessionExpiresAt sets the time at which the session expires.
func (wc *WebConn) SetSessionExpiresAt(v int64) {
	atomic.StoreInt64(&wc.sessionExpiresAt, v)
}

// GetSessionToken returns the session token of the connection.
func (wc *WebConn) GetSessionToken() string {
	return wc.sessionToken.Load().(string)
}

// SetSessionToken sets the session token of the connection.
func (wc *WebConn) SetSessionToken(v string) {
	wc.sessionToken.Store(v)
}

// GetSession returns the session of the connection.
func (wc *WebConn) GetSession() *model.Session {
	return wc.session.Load().(*model.Session)
}

// SetSession sets the session of the connection.
func (wc *WebConn) SetSession(v *model.Session) {
	if v != nil {
		v = v.DeepCopy()
	}

	wc.session.Store(v)
}

// Pump starts the WebConn instance. After this, the websocket
// is ready to send messages.
// This is only used by *nix platforms.
func (wc *WebConn) Pump() {
	// writePump is blocking in nature.
	wc.writePump()
	// Once it exits, we close everything.
	wc.App.HubUnregister(wc)
	close(wc.pumpFinished)
}

// BlockingPump is the Windows alternative of Pump.
// It creates two goroutines - one for reading, another
// for writing.
func (wc *WebConn) BlockingPump() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		wc.writePump()
	}()
	wc.readPump()
	close(wc.endWritePump)
	wg.Wait()
	wc.App.HubUnregister(wc)
	close(wc.pumpFinished)

	defer ReturnSessionToPool(wc.GetSession())
}

// startPoller adds the file descriptor of the connection
// to the global epoll instance and registers a callback.
func (wc *WebConn) startPoller() {
	desc := netpoll.Must(netpoll.HandleRead(wc.WebSocket))
	wc.App.Srv().Poller().Start(desc, func(wsEv netpoll.Event) {
		if wsEv&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			wc.App.Srv().Poller().Stop(desc)
			wc.Close()
			return
		}

		// Block until we have a token.
		wc.App.Srv().GetWebConnToken()
		// Read from conn.
		go func() {
			defer wc.App.Srv().ReleaseWebConnToken()
			err := wc.ReadMsg()
			if err != nil {
				mlog.Debug("Error while reading message from websocket", mlog.Err(err))
				wc.App.Srv().Poller().Stop(desc)
				// net.ErrClosed is not available until Go 1.16.
				// https://github.com/golang/go/issues/4373
				//
				// Sometimes, the netpoller generates a data event and a HUP event
				// close to each other. In that case, we don't want to double-close
				// the connection.
				if !strings.Contains(err.Error(), "use of closed network connection") {
					wc.Close()
				}
			}
		}()
	})
}

// GetWebConnToken creates backpressure by using
// a counting semaphore to limit the number of concurrent goroutines.
func (s *Server) GetWebConnToken() {
	s.webConnSemaWg.Add(1)
	s.webConnSema <- struct{}{}
}

// ReleaseWebConnToken releases a token
// got from the semaphore
func (s *Server) ReleaseWebConnToken() {
	<-s.webConnSema
	s.webConnSemaWg.Done()
}

// ReadMsg will read a single message from the websocket connection.
func (wc *WebConn) ReadMsg() error {
	r := wsutil.NewReader(wc.WebSocket, ws.StateServerSide)
	r.MaxFrameSize = model.SOCKET_MAX_MESSAGE_SIZE_KB
	decoder := json.NewDecoder(r)

	// The reader's methods are not goroutine safe.
	// We restrict only one reader goroutine per-connection.
	wc.readMut.Lock()
	defer wc.readMut.Unlock()

	var req model.WebSocketRequest
	hdr, err := r.NextFrame()
	if err != nil {
		return errors.Wrap(err, "error while getting the next websocket frame")
	}
	switch hdr.OpCode {
	case ws.OpClose:
		// Return if closed.
		// We need to return an error for Windows to let the reader exit.
		if wc.isWindows {
			return errors.New("connection closed")
		}
		return nil
	case ws.OpPong:
		wc.WebSocket.SetReadDeadline(time.Now().Add(pongWaitTime))
		// Handle pongs
		if wc.IsAuthenticated() {
			wc.App.Srv().Go(func() {
				wc.App.SetStatusAwayIfNeeded(wc.UserId, false)
			})
		}
	default:
		// Default case of data message.
		if err := decoder.Decode(&req); err != nil {
			// We discard any remaining data left in the socket.
			r.Discard()
			return errors.Wrap(err, "error during decoding websocket message")
		}

		wc.App.Srv().WebSocketRouter.ServeWebSocket(wc, &req)
	}
	return nil
}

func (wc *WebConn) readPump() {
	defer wc.WebSocket.Close()
	wc.WebSocket.SetReadDeadline(time.Now().Add(pongWaitTime))

	for {
		if err := wc.ReadMsg(); err != nil {
			wc.logSocketErr("websocket.read", err)
			return
		}
	}
}

func (wc *WebConn) writePump() {
	ticker := time.NewTicker(pingInterval)
	authTicker := time.NewTicker(authCheckInterval)

	defer func() {
		ticker.Stop()
		authTicker.Stop()
		wc.WebSocket.Close()
	}()

	var buf bytes.Buffer
	// 2k is seen to be a good heuristic under which 98.5% of message sizes remain.
	buf.Grow(1024 * 2)
	enc := json.NewEncoder(&buf)

	for {
		select {
		case msg, ok := <-wc.send:
			if !ok {
				wc.WebSocket.SetWriteDeadline(time.Now().Add(writeWaitTime))
				wsutil.WriteServerMessage(wc.WebSocket, ws.OpClose, []byte{})
				return
			}

			evt, evtOk := msg.(*model.WebSocketEvent)

			skipSend := false
			if len(wc.send) >= sendSlowWarn {
				// When the pump starts to get slow we'll drop non-critical messages
				switch msg.EventType() {
				case model.WEBSOCKET_EVENT_TYPING,
					model.WEBSOCKET_EVENT_STATUS_CHANGE,
					model.WEBSOCKET_EVENT_CHANNEL_VIEWED:
					mlog.Warn(
						"websocket.slow: dropping message",
						mlog.String("user_id", wc.UserId),
						mlog.String("type", msg.EventType()),
						mlog.String("channel_id", evt.GetBroadcast().ChannelId),
					)
					skipSend = true
				}
			}

			if skipSend {
				continue
			}

			buf.Reset()
			var err error
			if evtOk {
				cpyEvt := evt.SetSequence(wc.Sequence)
				err = cpyEvt.Encode(enc)
				wc.Sequence++
			} else {
				err = enc.Encode(msg)
			}
			if err != nil {
				mlog.Warn("Error in encoding websocket message", mlog.Err(err))
				continue
			}

			if len(wc.send) >= sendFullWarn {
				logData := []mlog.Field{
					mlog.String("user_id", wc.UserId),
					mlog.String("type", msg.EventType()),
					mlog.Int("size", buf.Len()),
				}
				if evtOk {
					logData = append(logData, mlog.String("channel_id", evt.GetBroadcast().ChannelId))
				}

				mlog.Warn("websocket.full", logData...)
			}

			wc.WebSocket.SetWriteDeadline(time.Now().Add(writeWaitTime))
			if err := wsutil.WriteServerMessage(wc.WebSocket, ws.OpText, buf.Bytes()); err != nil {
				wc.logSocketErr("websocket.send", err)
				return
			}

			if wc.App.Metrics() != nil {
				wc.App.Metrics().IncrementWebSocketBroadcast(msg.EventType())
			}
		case <-ticker.C:
			wc.WebSocket.SetWriteDeadline(time.Now().Add(writeWaitTime))
			if err := wsutil.WriteServerMessage(wc.WebSocket, ws.OpPing, []byte{}); err != nil {
				wc.logSocketErr("websocket.ticker", err)
				return
			}

		case <-wc.endWritePump:
			return

		case <-authTicker.C:
			if wc.GetSessionToken() == "" {
				mlog.Debug("websocket.authTicker: did not authenticate", mlog.Any("ip_address", wc.WebSocket.RemoteAddr()))
				return
			}
			authTicker.Stop()
		}
	}
}

// InvalidateCache resets all internal data of the WebConn.
func (wc *WebConn) InvalidateCache() {
	wc.allChannelMembers = nil
	wc.lastAllChannelMembersTime = 0
	wc.SetSession(nil)
	wc.SetSessionExpiresAt(0)
}

// IsAuthenticated returns whether the given WebConn is authenticated or not.
func (wc *WebConn) IsAuthenticated() bool {
	// Check the expiry to see if we need to check for a new session
	if wc.GetSessionExpiresAt() < model.GetMillis() {
		if wc.GetSessionToken() == "" {
			return false
		}

		session, err := wc.App.GetSession(wc.GetSessionToken())
		if err != nil {
			if err.StatusCode >= http.StatusBadRequest && err.StatusCode < http.StatusInternalServerError {
				mlog.Debug("Invalid session.", mlog.Err(err))
			} else {
				mlog.Error("Could not get session", mlog.String("session_token", wc.GetSessionToken()), mlog.Err(err))
			}

			wc.SetSessionToken("")
			wc.SetSession(nil)
			wc.SetSessionExpiresAt(0)
			return false
		}

		wc.SetSession(session)
		wc.SetSessionExpiresAt(session.ExpiresAt)
	}

	return true
}

func (wc *WebConn) createHelloMessage() *model.WebSocketEvent {
	msg := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_HELLO, "", "", wc.UserId, nil)
	msg.Add("server_version", fmt.Sprintf("%v.%v.%v.%v", model.CurrentVersion, model.BuildNumber, wc.App.ClientConfigHash(), wc.App.Srv().License() != nil))
	return msg
}

func (wc *WebConn) shouldSendEventToGuest(msg *model.WebSocketEvent) bool {
	var userID string
	var canSee bool

	switch msg.EventType() {
	case model.WEBSOCKET_EVENT_USER_UPDATED:
		user, ok := msg.GetData()["user"].(*model.User)
		if !ok {
			mlog.Debug("webhub.shouldSendEvent: user not found in message", mlog.Any("user", msg.GetData()["user"]))
			return false
		}
		userID = user.Id
	case model.WEBSOCKET_EVENT_NEW_USER:
		userID = msg.GetData()["user_id"].(string)
	default:
		return true
	}

	canSee, err := wc.App.UserCanSeeOtherUser(wc.UserId, userID)
	if err != nil {
		mlog.Error("webhub.shouldSendEvent.", mlog.Err(err))
		return false
	}

	return canSee
}

// shouldSendEvent returns whether the message should be sent or not.
func (wc *WebConn) shouldSendEvent(msg *model.WebSocketEvent) bool {
	// IMPORTANT: Do not send event if WebConn does not have a session
	if !wc.IsAuthenticated() {
		return false
	}

	// If the event contains sanitized data, only send to users that don't have permission to
	// see sensitive data. Prevents admin clients from receiving events with bad data
	var hasReadPrivateDataPermission *bool
	if msg.GetBroadcast().ContainsSanitizedData {
		hasReadPrivateDataPermission = model.NewBool(wc.App.RolesGrantPermission(wc.GetSession().GetUserRoles(), model.PERMISSION_MANAGE_SYSTEM.Id))

		if *hasReadPrivateDataPermission {
			return false
		}
	}

	// If the event contains sensitive data, only send to users with permission to see it
	if msg.GetBroadcast().ContainsSensitiveData {
		if hasReadPrivateDataPermission == nil {
			hasReadPrivateDataPermission = model.NewBool(wc.App.RolesGrantPermission(wc.GetSession().GetUserRoles(), model.PERMISSION_MANAGE_SYSTEM.Id))
		}

		if !*hasReadPrivateDataPermission {
			return false
		}
	}

	// If the event is destined to a specific user
	if msg.GetBroadcast().UserId != "" {
		return wc.UserId == msg.GetBroadcast().UserId
	}

	// if the user is omitted don't send the message
	if len(msg.GetBroadcast().OmitUsers) > 0 {
		if _, ok := msg.GetBroadcast().OmitUsers[wc.UserId]; ok {
			return false
		}
	}

	// Only report events to users who are in the channel for the event
	if msg.GetBroadcast().ChannelId != "" {
		if model.GetMillis()-wc.lastAllChannelMembersTime > webConnMemberCacheTime {
			wc.allChannelMembers = nil
			wc.lastAllChannelMembersTime = 0
		}

		if wc.allChannelMembers == nil {
			result, err := wc.App.Srv().Store.Channel().GetAllChannelMembersForUser(wc.UserId, true, false)
			if err != nil {
				mlog.Error("webhub.shouldSendEvent.", mlog.Err(err))
				return false
			}
			wc.allChannelMembers = result
			wc.lastAllChannelMembersTime = model.GetMillis()
		}

		if _, ok := wc.allChannelMembers[msg.GetBroadcast().ChannelId]; ok {
			return true
		}
		return false
	}

	// Only report events to users who are in the team for the event
	if msg.GetBroadcast().TeamId != "" {
		return wc.isMemberOfTeam(msg.GetBroadcast().TeamId)
	}

	if wc.GetSession().Props[model.SESSION_PROP_IS_GUEST] == "true" {
		return wc.shouldSendEventToGuest(msg)
	}

	return true
}

// IsMemberOfTeam returns whether the user of the WebConn
// is a member of the given teamID or not.
func (wc *WebConn) isMemberOfTeam(teamID string) bool {
	currentSession := wc.GetSession()

	if currentSession == nil || currentSession.Token == "" {
		session, err := wc.App.GetSession(wc.GetSessionToken())
		if err != nil {
			if err.StatusCode >= http.StatusBadRequest && err.StatusCode < http.StatusInternalServerError {
				mlog.Debug("Invalid session.", mlog.Err(err))
			} else {
				mlog.Error("Could not get session", mlog.String("session_token", wc.GetSessionToken()), mlog.Err(err))
			}
			return false
		}
		wc.SetSession(session)
		currentSession = session
	}

	return currentSession.GetTeamByTeamId(teamID) != nil
}

func (wc *WebConn) logSocketErr(source string, err error) {
	mlog.Debug(source+": error during writing to websocket", mlog.String("user_id", wc.UserId), mlog.Err(err))
}
