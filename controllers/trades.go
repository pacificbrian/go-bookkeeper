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

func CreateTrade(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Println("CREATE TRADE")
	db := gormdb.DbManager()

	entry := new(model.Trade)
	c.Bind(entry)
	entry.AccountID = uint(id)
	entry.Date = getFormDate(c)
	entry.Create(db)

	// how to determine if referred here by Account or Security?
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func DeleteTrade(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("DELETE TRADE(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Trade)
	entry.ID = uint(id)
	if entry.Delete(db) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func UpdateTrade(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("UPDATE TRADE(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Trade)
	entry.ID = uint(id)
	entry = entry.Get(db)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	c.Bind(entry)
	entry.Update(db)
	a_id := entry.AccountID
	// how to determine if referred here by Account or Security?
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", a_id))
	//return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/trade/%d", id))
}

func EditTrade(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("EDIT TRADE(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Trade)
	entry.ID = uint(id)
	entry = entry.Get(db)

	data := map[string]any{ "trade": entry,
				"trade_types": new(model.TradeType).List(db) }
	return c.Render(http.StatusOK, "trade/edit.html", data)
}
