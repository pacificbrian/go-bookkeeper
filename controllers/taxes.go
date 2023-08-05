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
	"github.com/pacificbrian/go-bookkeeper/helpers"
	"github.com/pacificbrian/go-bookkeeper/model"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func ListTaxes(c echo.Context) error {
	year, _ := strconv.Atoi(c.Param("year"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("LIST TAXES (%d)", year)
	get_json := false

	returns := new(model.TaxReturn).List(session, year)

	dh := new(helpers.DateHelper)
	dh.Init()
	if year > 0 {
		dh.SetYear(year)
	}

	if get_json {
		// TODO: add separate ListTaxEntries if needed
		return c.JSON(http.StatusOK, returns)
	} else {
		db := session.DB
		entry := new(model.TaxEntry)
		entries := new(model.TaxEntry).List(session, year)
		data := map[string]any{ "tax_returns": returns,
					"tax_entries": entries,
					"date_helper": dh,
					"entry": entry,
					"filing_status": new(model.TaxFilingStatus).List(db),
					"isEdit": false,
					"tax_items": new(model.TaxItem).List(db),
					"tax_regions": new(model.TaxRegion).List(db),
					"tax_types": new(model.TaxType).List(db),
					"account": nil,
					"year": year }
		return c.Render(http.StatusOK, "taxes/index.html", data)
	}
}

func ListTaxCashFlows(c echo.Context) error {
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	var entries []model.CashFlow
	taxItem := new(model.TaxItem)
	taxType := new(model.TaxType)
	get_json := false

	year, _ := strconv.Atoi(c.Param("year"))
	paramID, _ := strconv.Atoi(c.Param("tax_item_id"))
	taxItem.ID = uint(paramID)
	paramID, _ = strconv.Atoi(c.Param("tax_type_id"))
	taxType.ID = uint(paramID)

	log.Printf("LIST TAX CASHFLOWS (%d) ITEM(%d) TYPE(%d)",
		   year, taxItem.ID, taxType.ID)


	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		db := session.DB
		taxItem  = taxItem.Get(db)
		taxType  = taxType.Get(db)
		if taxItem != nil && taxItem.ID > 0 {
			entries,_ = taxItem.ListTaxCashFlows(session, year)
		} else if taxType != nil && taxType.ID > 0 {
			entries,_ = taxType.ListTaxCashFlows(session, year)
		}

		data := map[string]any{ "cash_flows": entries,
					"tax_item": taxItem,
					"tax_type": taxType,
					"account": nil,
					"disallow_cashflow_delete": true,
					"year": year }
		return c.Render(http.StatusOK, "taxes/list_entries.html", data)
	}
}

func CreateTaxEntry(c echo.Context) error {
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Println("CREATE TAX ENTRY")

	entry := new(model.TaxEntry)
	err := c.Bind(entry)
	assert(err == nil, "CREATE TAX ENTRY BIND FAILED")
	entry.Amount = getFormDecimal(c, "amount")
	entry.Create(session)

	// http.StatusCreated
	if entry.DateYear > 0 {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/years/%d/taxes", entry.DateYear))
	} else {
		return c.Redirect(http.StatusSeeOther, "/taxes")
	}
}

func DeleteTaxEntry(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("DELETE TAX ENTRY(%d)", id)

	entry := new(model.TaxEntry)
	entry.ID = uint(id)
	if entry.Delete(session) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func UpdateTaxEntry(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("UPDATE TAX ENTRY(%d)", id)

	entry := new(model.TaxEntry)
	entry.ID = uint(id)
	entry = entry.Get(session)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	err := c.Bind(entry)
	assert(err == nil, "UPDATE TAX ENTRY BIND FAILED")
	entry.Amount = getFormDecimal(c, "amount")
	entry.Update()
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/years/%d/taxes",entry.DateYear))
}

func EditTaxEntry(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("EDIT TAX ENTRY(%d)", id)

	entry := new(model.TaxEntry)
	entry.ID = uint(id)
	entry = entry.Get(session)

	data := map[string]any{ "entry": entry,
				"isEdit": true,
				"tax_items": new(model.TaxItem).List(session.DB),
				"tax_types": new(model.TaxType).List(session.DB),
				"tax_regions": new(model.TaxRegion).List(session.DB) }
	return c.Render(http.StatusOK, "taxes/edit.html", data)
}

func CreateTaxes(c echo.Context) error {
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Println("CREATE TAX RETURN")

	entry := new(model.TaxReturn)
	c.Bind(entry)
	entry.Create(session)

	// http.StatusCreated
	if entry.Year > 0 {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/years/%d/taxes", entry.Year))
	} else {
		return c.Redirect(http.StatusSeeOther, "/taxes")
	}
}

func RecalculateTaxes(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("RECALCULATE TAX RETURN(%d)", id)

	entry := new(model.TaxReturn)
	entry.ID = uint(id)

	// Recalculate will verify if have TaxReturn access
	if entry.Recalculate(session) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func DeleteTaxes(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("DELETE TAX RETURN(%d)", id)

	entry := new(model.TaxReturn)
	entry.ID = uint(id)
	if entry.Delete(session) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}
