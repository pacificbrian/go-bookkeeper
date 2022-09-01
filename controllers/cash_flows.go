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

func CreateCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("CREATE CASHFLOW (ACCOUNT:%d)", id)
	db := gormdb.DbManager()

	entry := new(model.CashFlow)
	entry.AccountID = uint(id)
	c.Bind(entry)
	spew.Dump(entry)

	db.Create(entry)

	// http.StatusCreated
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func DeleteCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("DELETE CASHFLOW(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	spew.Dump(entry)
	db.Delete(entry)

	return c.NoContent(http.StatusAccepted)
}

func UpdateCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("UPDATE CASHFLOW(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	db.First(&entry)
	// verify entry id was valid

	c.Bind(entry)
	spew.Dump(entry)
	//db.Save(entry)

	a_id := entry.Account.ID
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", a_id))
}

func EditCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("EDIT CASHFLOW(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	db.First(&entry)

	data := map[string]any{ "cash_flow": entry,
				"button_text": "Update" }
	return c.Render(http.StatusOK, "cash_flows/edit.html", data)
}
