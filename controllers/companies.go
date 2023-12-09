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

const maxFilings = 3

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func GetCompanyFinancials(c echo.Context) error {
	limitFilings, _ := strconv.Atoi(c.QueryParam("limit"))
	filingType := c.QueryParam("type")
	id, _ := strconv.Atoi(c.Param("company_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	get_json := false

        entry := new(model.Company)
        entry = entry.Find(uint(id))
        if entry == nil {
                return c.NoContent(http.StatusUnauthorized)
        }

	if filingType == "" {
		filingType = "10-K"
	}
	var filings []model.FilingData
	filingDates := entry.GetFilingDates(filingType)
	viewName := model.ConsolidatedStatements
	filingItemNames := entry.GetFilingItemNames(viewName)
	log.Printf("LIST COMPANY(%d) STATEMENT(%s:%s:%d)", id, viewName,
		   filingType, len(filingDates))

	if limitFilings == 0 {
		limitFilings = maxFilings
	}
	for _, date := range filingDates {
		f := entry.GetFiling(filingType, date)
		if f != nil {
			filings = append(filings, f)
			if len(filings) >= limitFilings {
				break
			}
		}
	}

	if get_json {
		return c.JSON(http.StatusOK, filings)
	} else {
		data := map[string]any{ "company": entry,
					"filings": filings,
					"filingItemNames": filingItemNames,
					"filingViewName": viewName }
		return c.Render(http.StatusOK,
				"companies/financials.html", data)
	}
}
