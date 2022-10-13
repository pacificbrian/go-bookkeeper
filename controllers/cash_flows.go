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
	"go-bookkeeper/helpers"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func CreateCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("CREATE CASHFLOW (ACCOUNT:%d)", id)
	db := gormdb.DebugDbManager()

	entry := new(model.CashFlow)
	c.Bind(entry)
	entry.AccountID = uint(id)
	entry.Date = getFormDate(c)
	entry.Create(db)

	// http.StatusCreated
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func CreateScheduledCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("CREATE SCHEDULED CASHFLOW (FROM:%d)", id)
	db := gormdb.DebugDbManager()

	entry := new(model.CashFlow)
	c.Bind(entry)
	c.Bind(&entry.RepeatInterval)
	entry.AccountID = uint(id)
	entry.Date = getFormDate(c)
	entry.Type = "RCashFlow"
	entry.Create(db)

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func CreateSplitCashFlow(c echo.Context) error {
	split_from, _ := strconv.Atoi(c.Param("id"))
	log.Printf("CREATE SPLIT CASHFLOW (PARENT:%d)", split_from)
	db := gormdb.DebugDbManager()

	entry, httpStatus := model.NewSplitCashFlow(db, uint(split_from))
	if entry == nil {
		return c.NoContent(httpStatus)
	}

	// from NewSplitCashFlow, we already have: AccountID, Date, Payee
	c.Bind(entry)
	entry.Create(db)
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/cash_flows/%d/edit", split_from))
}

func PutCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("PUT CASHFLOW(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)

	kv := new(PutKeyValue)
	c.Bind(kv)
	putRequest := make(map[string]interface{})
	putRequest[kv.Key] = kv.Value

	if entry.Put(db, putRequest) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
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
	db := gormdb.DebugDbManager()

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	entry = entry.Get(db, false)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	c.Bind(entry)
	c.Bind(&entry.RepeatInterval)
	entry.Update(db)
	a_id := entry.AccountID
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", a_id))
}

func EditCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("EDIT CASHFLOW(%d)", id)
	db := gormdb.DbManager()

	var cash_flows []model.CashFlow
	var cash_flow_total string
	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	entry = entry.Get(db, true)
	if entry != nil {
		cash_flows, cash_flow_total = entry.ListSplit(db)
	}

	dh := new(helpers.DateHelper)
	dh.Init()
	dh.SetDate(entry.Date)

	data := map[string]any{ "cash_flow": entry,
				"date_helper": dh,
				"cash_flows": cash_flows,
				"total_amount": cash_flow_total,
				"cash_flow_types": new(model.CashFlowType).List(db),
				"categories": new(model.Category).List(db) }
	return c.Render(http.StatusOK, "cash_flows/edit.html", data)
}

func ListScheduledCashFlows(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("LIST SCHEDULED CASHFLOWS (ACCOUNT:%d)", id)
	db := gormdb.DebugDbManager()

	var cash_flows []model.CashFlow
	entry := new(model.CashFlow)
	entry.AccountID = uint(id)
	entry.Type = "RCashFlow"

	entry.Account.ID = entry.AccountID
	cash_flows = entry.Account.ListScheduled(db, false)

	dh := new(helpers.DateHelper)
	dh.Init()

	data := map[string]any{ "cash_flow": entry,
				"date_helper": dh,
				"button_text": "Add Scheduled",
				"cash_flows": cash_flows,
				"cash_flow_types": new(model.CashFlowType).List(db),
				"repeat_interval_types": new(model.RepeatIntervalType).List(db),
				"categories": new(model.Category).List(db) }
	return c.Render(http.StatusOK, "cash_flows/index.html", data)
}
