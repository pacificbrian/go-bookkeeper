/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/pacificbrian/go-bookkeeper/model"
)

func Login(c echo.Context) error {
	if !IsEnabledMultiUser() {
		return CreateSession(c)
	}

	data := map[string]any{ "user": new(model.User) }
	return c.Render(http.StatusOK, "users/login.html", data)
}
