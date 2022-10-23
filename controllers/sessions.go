/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

func Login(c echo.Context) error {
	return c.Redirect(http.StatusSeeOther, "/accounts")
}
