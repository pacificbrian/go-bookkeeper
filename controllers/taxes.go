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
	"go-bookkeeper/helpers"
	"go-bookkeeper/model"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func ListTaxes(c echo.Context) error {
	year, _ := strconv.Atoi(c.Param("year"))
	log.Printf("LIST TAXES (%d)", year)
	db := gormdb.DbManager()
	get_json := false

	returns := new(model.TaxReturn).List(db)

	dh := new(helpers.DateHelper)
	dh.Init()
	if year > 0 {
		dh.SetYear(year)
	}

	if get_json {
		// TODO: add separate ListTaxEntries if needed
		return c.JSON(http.StatusOK, returns)
	} else {
		entries := new(model.TaxEntry).List(db)
		data := map[string]any{ "tax_returns": returns,
					"tax_entries": entries,
					"date_helper": dh,
					"filing_status": new(model.TaxFilingStatus).List(db),
					"tax_items": new(model.TaxItem).List(db),
					"tax_regions": new(model.TaxRegion).List(db),
					"tax_types": new(model.TaxType).List(db),
					"account": nil,
					"year": year }
		return c.Render(http.StatusOK, "taxes/index.html", data)
	}
}

func CreateTaxEntry(c echo.Context) error {
	log.Println("CREATE TAX ENTRY")
	db := gormdb.DbManager()

	entry := new(model.TaxEntry)
	c.Bind(entry)
	entry.Create(db)

	// http.StatusCreated
	if entry.DateYear > 0 {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/years/%d/taxes", entry.DateYear))
	} else {
		return c.Redirect(http.StatusSeeOther, "/taxes")
	}
}

func CreateTaxes(c echo.Context) error {
	log.Println("CREATE TAX RETURN")
	db := gormdb.DbManager()

	entry := new(model.TaxReturn)
	c.Bind(entry)
	entry.Create(db)

	// http.StatusCreated
	if entry.Year > 0 {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/years/%d/taxes", entry.Year))
	} else {
		return c.Redirect(http.StatusSeeOther, "/taxes")
	}
}
