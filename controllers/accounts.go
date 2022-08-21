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
	"github.com/davecgh/go-spew/spew"
	gormdb "go-bookkeeper/db"
	"go-bookkeeper/model"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

// Need Access Controls
// Test with InPlace POST (AJAX is now FetchAPI ?)
// Use Controller objects
// Need Input Validation after Bind

func ListAccounts(c echo.Context) error {
	log.Println("LIST ACCOUNTS")
	db := gormdb.DbManager()
	get_json := false

	entries := model.ListAccounts(db)
	//spew.Dump(entries)

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		data := map[string]any{ "accounts":entries }
		return c.Render(http.StatusOK, "accounts/index.html", data)
	}
}

func CreateAccount(c echo.Context) error {
	log.Println("CREATE ACCOUNT")
	db := gormdb.DbManager()

	entry := new(model.Account)
	c.Bind(entry)
	spew.Dump(entry)

	db.Create(entry)

	return c.Redirect(http.StatusSeeOther, "/accounts")
}

func DeleteAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("DELETE ACCOUNT(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	spew.Dump(entry)
	db.Delete(entry)

	return c.NoContent(http.StatusAccepted)
}

func GetAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("GET ACCOUNT(%d)", id)
	db := gormdb.DbManager()
	get_json := false

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	db.First(&entry)

	spew.Dump(entry)

	if get_json {
		return c.JSON(http.StatusOK, entry)
	} else {
		data := map[string]any{ "account":entry }
		return c.Render(http.StatusOK, "accounts/show.html", data)
	}
}

func UpdateAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	db := gormdb.DbManager()

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	db.First(&entry)
	// verify entry id was valid

	c.Bind(entry)
	spew.Dump(entry)
	//db.Save(entry)

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func EditAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("EDIT ACCOUNT(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	db.First(&entry)

	data := map[string]any{ "account": entry,
				"button_text": "Update Account",
				"account_types": new(model.AccountType).List(db),
				"currency_types": new(model.CurrencyType).List(db) }
	return c.Render(http.StatusOK, "accounts/edit.html", data)
}

func NewAccount(c echo.Context) error {
	log.Println("NEW ACCOUNT")
	db := gormdb.DbManager()

	data := map[string]any{ "account": new(model.Account),
				"button_text": "Create Account",
				"account_types": new(model.AccountType).List(db),
				"currency_types": new(model.CurrencyType).List(db) }
	return c.Render(http.StatusOK, "accounts/new.html", data)
}
