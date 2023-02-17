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
	"github.com/pacificbrian/go-bookkeeper/model"
	"github.com/pacificbrian/go-bookkeeper/helpers"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func CreateCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("CREATE CASHFLOW ACCOUNT(%d)", id)

	entry := new(model.CashFlow)
	err := c.Bind(entry)
	assert(err == nil, "CREATE CASHFLOW BIND FAILED")
	entry.AccountID = uint(id)
	entry.Amount = getFormDecimal(c, "amount")
	entry.Date = getFormDate(c)
	err = entry.Create(session)
	if err != nil {
		log.Printf("CREATE CASHFLOW ACCOUNT(%d) FAILED: %v", id, err)
	}

	// http.StatusCreated
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func CreateScheduledCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("CREATE SCHEDULED CASHFLOW ACCOUNT(%d)", id)

	entry := new(model.CashFlow)
	err := c.Bind(entry)
	assert(err == nil, "CREATE SCHEDULED CASHFLOW BIND FAILED")
	c.Bind(&entry.RepeatInterval)
	entry.RepeatInterval.Rate = getFormDecimal(c, "rate")
	entry.AccountID = uint(id)
	entry.Date = getFormDate(c)
	entry.Amount = getFormDecimal(c, "amount")
	entry.Type = "RCashFlow"
	err = entry.Create(session)
	if err != nil {
		log.Printf("CREATE SCHEDULED CASHFLOW ACCOUNT(%d) FAILED: %v", id, err)
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d/scheduled", id))
}

func CreateSplitCashFlow(c echo.Context) error {
	split_from, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("CREATE SPLIT CASHFLOW (PARENT:%d)", split_from)

	entry, httpStatus := model.NewSplitCashFlow(session, uint(split_from))
	if entry == nil {
		return c.NoContent(httpStatus)
	}

	// from NewSplitCashFlow, we already have: AccountID, Date, Payee
	err := c.Bind(entry)
	assert(err == nil, "CREATE SPLIT CASHFLOW BIND FAILED")
	entry.Amount = getFormDecimal(c, "amount")
	err = entry.Create(session)
	if err != nil {
		log.Printf("CREATE SPLIT CASHFLOW FAILED: %v", err)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/cash_flows/%d/edit", split_from))
}

func PutCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("PUT CASHFLOW(%d)", id)

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)

	kv := new(PutKeyValue)
	c.Bind(kv)
	putRequest := make(map[string]interface{})
	putRequest[kv.Key] = kv.Value

	// Put will verify if have CashFlow access
	if entry.Put(session, putRequest) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func DeleteCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("DELETE CASHFLOW(%d)", id)

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	if entry.Delete(session) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func UpdateCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("UPDATE CASHFLOW(%d)", id)

	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	entry = entry.Get(session, false)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	err := c.Bind(entry)
	assert(err == nil, "UPDATE CASHFLOW BIND FAILED")
	c.Bind(&entry.RepeatInterval)
	entry.RepeatInterval.Rate = getFormDecimal(c, "rate")
	entry.Amount = getFormDecimal(c, "amount")
	entry.Date = getFormDate(c)
	// special case RepeatsLeft so that unset from user equals SQL NULL value
	entry.RepeatInterval.SetRepeatsLeft(c.FormValue("repeats"))
	entry.Update()

	// possibly can clean this up with Sessions
	if entry.Split {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/cash_flows/%d/edit",
								   entry.ParentID()))
	} else if entry.IsScheduled() {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d/scheduled",
								   entry.AccountID))
	} else {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d",
								   entry.AccountID))
	}
}

func EditCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("EDIT CASHFLOW(%d)", id)
	db := session.DB

	var repeat_interval_types []model.RepeatIntervalType
	var cash_flows []model.CashFlow
	var cash_flow_total string
	entry := new(model.CashFlow)
	entry.Model.ID = uint(id)
	entry = entry.Get(session, true)
	if entry != nil {
		cash_flows, cash_flow_total = entry.ListSplit(db)
		if entry.IsScheduled() {
			repeat_interval_types = new(model.RepeatIntervalType).List(db)
		}
	}

	dh := new(helpers.DateHelper)
	dh.Init()
	dh.SetDate(entry.Date)

	data := map[string]any{ "cash_flow": entry,
				"date_helper": dh,
				"cash_flows": cash_flows,
				"total_amount": cash_flow_total,
				"cash_flow_types": new(model.CashFlowType).List(db),
				"categories": new(model.Category).List(db),
				"repeat_interval_types": repeat_interval_types }
	return c.Render(http.StatusOK, "cash_flows/edit.html", data)
}

func ListScheduledCashFlows(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("LIST SCHEDULED CASHFLOWS (ACCOUNT:%d)", id)
	db := session.DB

	var cash_flows []model.CashFlow
	entry := new(model.CashFlow)
	entry.AccountID = uint(id)
	entry.Type = "RCashFlow"

	entry.Account.ID = entry.AccountID
	cash_flows = entry.Account.ListScheduled(session, false)

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
