/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"gorm.io/gorm"
)

type Company struct {
	Model
	Name string `form:"Name"`
	Symbol string `form:"Symbol"`
}

func (co Company) GetName() string {
	if co.Name != "" {
		return co.Name
	}
	return co.Symbol
}

func companyGetBySymbol(db *gorm.DB, symbol string) *Company {
	company := new(Company)
	company.Symbol = symbol
	// need Where because these are not primary keys
	db.Where(&company).First(&company)

	if company.ID == 0 {
		db.Create(company)
		spewModel(company)
		log.Printf("[MODEL] CREATE COMPANY(%d)", company.ID)
	}

	return company
}
