/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"fmt"
	"log"
	"time"
	"net/http"
	"unsafe"
	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/pacificbrian/go-bookkeeper/config"
	"github.com/pacificbrian/go-bookkeeper/model"
)

const SessionExpiry = 3
var sessionManager *scs.SessionManager
var activeSessions map[uint]*model.Session
var defaultSession *model.Session

// could have separate config variable for multi-user, but as this is what
// sessions are used to managed, just tie to if sessions are enabled.
func IsEnabledMultiUser() bool {
	return sessionManager != nil
}

func StartSessionManager() *scs.SessionManager {
	if !config.GlobalConfig().Sessions {
		return nil
	}
	sessionManager = scs.New()
	//sessionManager.Codec = PointerCodec{}
	sessionManager.Lifetime = SessionExpiry * 24 * time.Hour
	return sessionManager
}

func getSession(c echo.Context) *model.Session {
	if sessionManager == nil {
		return defaultSession
	}

	msg := sessionManager.GetString(c.Request().Context(), "tag")
	uID := sessionManager.GetInt(c.Request().Context(), "user_id")
	sessionPtr, valid := sessionManager.Get(c.Request().Context(), "session").(uintptr)
	if !valid {
		log.Printf("GET SESSION ERROR from sessionManager.Get")
		return nil
	}

	userSession := (*model.Session)(unsafe.Pointer(sessionPtr))
	sessionUserID := int(userSession.User.ID)
	if !(uID > 0 && uID == sessionUserID) {
		log.Printf("GET SESSION NOT FOUND")
		return nil
	}

	log.Printf("GET SESSION FOR([%d,%d]) FOUND TAG(%s)", uID, sessionUserID, msg)
	return userSession
}

func newSession(c echo.Context, u *model.User) {
	if sessionManager == nil {
		return
	}
	msg := fmt.Sprintf("[%d] created at: %s", u.ID, timeToString(time.Now()))
	sessionManager.Put(c.Request().Context(), "tag", msg)
	sessionManager.Put(c.Request().Context(), "user_id", int(u.ID))
	// Creation session and store in map so not garbage collected.
	// Note, the goal was to not have to maintain activeSessions and only
	// store pointer directly in the SessionData.
	// But unsafe.Pointer below can be garbage collected without this.
	// Then we don't actually need to put a pointer in the SessionData, as
	// we can just index in this map with sessionData["user_id"], but I
	// keep the logic below to use unsafe.Pointer just to remember this
	// usage of pointers as it took time to research this.
	userSession := u.NewSession()
	activeSessions[u.ID] = userSession

	// Don't want to flatten Session, store Session pointer in sessionData
	// Why is Go afraid of pointers?
	userSessionPtr := (uintptr)(unsafe.Pointer(userSession))
	sessionManager.Put(c.Request().Context(), "session", userSessionPtr)
	log.Printf("CREATE NEW SESSION FOR(%d) TAG(%s)", u.ID, msg)
}

func init() {
	activeSessions = map[uint]*model.Session{}
	defaultUser := new(model.User)
	defaultUser.ID = 1
	defaultSession = defaultUser.NewSession()
}

func redirectToLogin(c echo.Context) error {
	return c.Redirect(http.StatusSeeOther, "/")
}

// run via signal.ContextNotify from main
func CloseActiveSessions() {
	if sessionManager == nil {
		defaultSession.CloseSession()
		return
	}

	for _,v := range activeSessions {
		v.CloseSession()
	}
}

func CreateSession(c echo.Context) error {
	authenticated := false
	user := new(model.User)

	if IsEnabledMultiUser() {
		user = user.GetByLogin(c.FormValue("user.Login"))
		authenticated = user != nil &&
				user.Authenticate(c.FormValue("user.Password"))
	} else {
		user = defaultSession.GetUser()
		authenticated = true
	}

	if authenticated {
		newSession(c, user)
		return c.Redirect(http.StatusSeeOther, "/accounts")
	} else {
		return c.NoContent(http.StatusUnauthorized)
	}
}
