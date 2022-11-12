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
	oldName string `gorm:"-:all"`
	oldSymbol string `gorm:"-:all"`
}

func (co Company) GetName() string {
	if co.Name != "" {
		return co.Name
	}
	return co.Symbol
}

func companyGetBySymbol(db *gorm.DB, symbol string, name string) *Company {
	company := new(Company)
	company.Symbol = symbol
	company.Name = name
	// need Where because these are not primary keys
	db.Where(&company).First(&company)

	if company.ID == 0 {
		db.Create(company)
		spewModel(company)
		log.Printf("[MODEL] CREATE COMPANY(%d)", company.ID)
	}

	return company
}

func (c *Company) updateName(db *gorm.DB) error {
	var err error

	if c.oldSymbol == c.Symbol &&
	   c.oldName != c.Name {
		result := db.Model(c).Update("Name", c.Name)
		err = result.Error
	}
	return err
}

func (c *Company) update(db *gorm.DB) bool {
	if c.oldSymbol != c.Symbol {
		newCompany := companyGetBySymbol(db, c.Symbol, c.Name)
		c.ID = newCompany.ID
	}
	return c.oldSymbol != c.Symbol
}
