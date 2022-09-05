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
	"strconv"
	"github.com/labstack/echo/v4"
	gormdb "go-bookkeeper/db"
	"go-bookkeeper/model"
	"go-bookkeeper/helpers"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func CreateCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("CREATE CASHFLOW (ACCOUNT:%d)", id)
	db := gormdb.DbManager()

	entry := new(model.CashFlow)
	entry.AccountID = uint(id)
	date := c.FormValue("date_month") + "/" +
		c.FormValue("date_day") + "/" +
		c.FormValue("date_year")
	entry.Date,_ = time.Parse("1/2/2006", date)
	c.Bind(entry)
	entry.Create(db)

	// http.StatusCreated
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func DeleteCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("DELETE CASHFLOW(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	if entry.Delete(db) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func UpdateCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("UPDATE CASHFLOW(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	entry = entry.Get(db, false)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	c.Bind(entry)
	entry.Update(db)
	a_id := entry.AccountID
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", a_id))
}

func EditCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("EDIT CASHFLOW(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	entry = entry.Get(db, true)

	dh := new(helpers.DateHelper)
	dh.Init()

	data := map[string]any{ "cash_flow": entry,
				"date_helper": dh,
				"cash_flow_types": new(model.CashFlowType).List(db),
				"categories": new(model.CategoryType).List(db) }
	return c.Render(http.StatusOK, "cash_flows/edit.html", data)
}
