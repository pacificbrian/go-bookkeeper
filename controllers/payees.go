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

func ListPayees(c echo.Context) error {
	log.Println("LIST PAYEES")
	db := gormdb.DbManager()
	get_json := false

	entries := new(model.Payee).List(db)

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		data := map[string]any{ "payees":entries,
					"account":nil }
		return c.Render(http.StatusOK, "payees/index.html", data)
	}
}

func CreatePayee(c echo.Context) error {
	log.Println("CREATE PAYEE")
	db := gormdb.DbManager()

	entry := new(model.Payee)
	c.Bind(entry)
	entry.Create(db)

	// http.StatusCreated
	return c.Redirect(http.StatusSeeOther, "/payees")
}

func DeletePayee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("DELETE PAYEE(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Payee)
	entry.ID = uint(id)
	entry.Delete(db)
	// set status based on if Delete failed
	// return c.NoContent(http.StatusUnauthorized)

	return c.NoContent(http.StatusAccepted)
}

func UpdatePayee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("UPDATE PAYEE(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Payee)
	entry.ID = uint(id)
	entry = entry.Get(db)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	c.Bind(entry)
	entry.Update(db)
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/payee/%d", id))
}

func EditPayee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("EDIT PAYEE(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Payee)
	entry.ID = uint(id)
	entry = entry.Get(db)

	data := map[string]any{ "payee": entry,
				"categories": new(model.CategoryType).List(db) }
	return c.Render(http.StatusOK, "payee/edit.html", data)
}
