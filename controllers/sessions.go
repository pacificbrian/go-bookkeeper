/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/pacificbrian/go-bookkeeper/model"
)

var defaultSession *model.Session

func GetCurrentUser() *model.User {
	return &defaultSession.User
}

func getSession(c echo.Context) *model.Session {
	return defaultSession
}

func init() {
	defaultUser := new(model.User)
	defaultUser.ID = 1
	defaultSession = defaultUser.NewSession()
}

func Login(c echo.Context) error {
	return c.Redirect(http.StatusSeeOther, "/accounts")
}
