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
)

func CreateImportedCashFlows(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}

	var importFile model.HttpFile
	file, err := c.FormFile("filename")
	if err == nil {
		log.Printf("IMPORT CASHFLOWS (ACCOUNT:%d) (FILE:%s)", id, file.Filename)
		importFile.FileName = file.Filename
		importFile.FileData, err = file.Open()
	}
	if err != nil {
		log.Println(err)
		return c.NoContent(http.StatusNoContent)
	}
	defer importFile.FileData.Close()

	entry := new(model.Import)
	entry.AccountID = uint(id)
	err = entry.ImportFile(session, importFile)
	if err != nil {
		log.Println(err)
		return c.NoContent(http.StatusNoContent)
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d/imported", id))
}

func ListImportedCashFlows(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("LIST IMPORTED CASHFLOWS (ACCOUNT:%d)", id)

	var imports []model.Import
	entry := new(model.Account)
	entry.ID = uint(id)
	imports = entry.ListImports(session)

	data := map[string]any{ "account": entry,
				"button_text": "Import File",
				"imports": imports }
	return c.Render(http.StatusOK, "accounts/import.html", data)
}
