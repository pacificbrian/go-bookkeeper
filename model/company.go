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

func (c Company) GetName() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Symbol
}

func (c *Company) Get(db *gorm.DB) *Company {
	if c.Name == "" && c.Symbol == "" {
		return nil
	}
	// need Where because these are not primary keys
	db.Where(c).First(c)

	if c.ID == 0 {
		db.Create(c)
		spewModel(c)
		log.Printf("[MODEL] CREATE COMPANY(%d)", c.ID)
	}

	return c
}

func companyGetBySymbol(db *gorm.DB, symbol string, name string) *Company {
	company := new(Company)
	company.Symbol = symbol
	company.Name = name
	return company.Get(db)
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

/* return true if Symbol was updated */
func (c *Company) update(db *gorm.DB) bool {
	if c.oldSymbol == c.Symbol {
		return false
	}

	newCompany := companyGetBySymbol(db, c.Symbol, c.Name)
	if newCompany == nil {
		return false
	}
	c.ID = newCompany.ID
	return true
}
