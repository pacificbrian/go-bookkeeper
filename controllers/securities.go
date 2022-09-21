/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"github.com/labstack/echo/v4"
	gormdb "go-bookkeeper/db"
	"go-bookkeeper/model"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func ListSecurities(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Println("LIST SECURITIES")
	db := gormdb.DbManager()
	get_json := false

	var entries []model.Security
	entry := new(model.Security)
	entry.AccountID = uint(id)
	entries = entry.List(db)

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		data := map[string]any{ "securities":entries,
					"account":&entry.Account }
		return c.Render(http.StatusOK, "securities/index.html", data)
	}
}

func CreateSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Println("CREATE SECURITY")
	db := gormdb.DbManager()

	entry := new(model.Security)
	c.Bind(entry)
	entry.AccountID = uint(id)
	entry.Create(db)

	// http.StatusCreated
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func DeleteSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("DELETE SECURITY(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Security)
	entry.ID = uint(id)
	if entry.Delete(db) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func UpdateSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("UPDATE SECURITY(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Security)
	entry.ID = uint(id)
	entry = entry.Get(db)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	c.Bind(entry)
	entry.Update(db)
	a_id := entry.AccountID
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", a_id))
	//return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/security/%d", id))
}

func EditSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("EDIT SECURITY(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Security)
	entry.ID = uint(id)
	entry = entry.Get(db)

	data := map[string]any{ "security": entry,
				"security_types": new(model.SecurityType).List(db) }
	return c.Render(http.StatusOK, "security/edit.html", data)
}
