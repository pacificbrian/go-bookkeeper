/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"net/http"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func userIDFromToken(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["id"].(string)
}

func Login(c echo.Context) error {
	return c.Redirect(http.StatusSeeOther, "/accounts")
}
