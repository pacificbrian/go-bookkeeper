/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"log"
	"net/http"
	"strconv"
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

func CreateUser(c echo.Context) error {
	var password [2]string

	entry := new(model.User)
	c.Bind(entry)
	log.Printf("CREATE USER LOGIN(%s)", entry.Login)

	password[0] = c.FormValue("user.Password")
	password[1] = c.FormValue("user.PasswordConfirmation")
	err := entry.Create(password)
	if err != nil {
		log.Printf("CREATE USER FAILED: %v", err)
		return c.Redirect(http.StatusSeeOther, "/users/new")
	}
	return c.Redirect(http.StatusSeeOther, "/accounts")
}

func DeleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("DELETE USER(%d)", id)

	entry := new(model.User)
	entry.ID = uint(id)
	if entry.Delete(session) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func NewUser(c echo.Context) error {
	log.Println("NEW USER")

	data := map[string]any{ "user": new(model.User) }
	return c.Render(http.StatusOK, "users/new.html", data)
}
