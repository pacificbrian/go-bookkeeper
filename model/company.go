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

// Query if Company already exists and return.
// If non-existent, create one. But if this is an Update, first
// try to update the calling object Company first if only attached
// to one Security.
// XXX Possibly should disallow multiple Companies for same Symbol.
// Or in future have defined set of fixed Companies if ever consider
// to add support for financial statements.
func (c *Company) Get(db *gorm.DB, tryUpdate bool) *Company {
	if c.Name == "" && c.Symbol == "" {
		return nil
	}
	result := new(Company)

	// Where query will only look at non-primary-key fields
	db.Where(c).First(result)

	if result.ID > 0 {
		c = result
		log.Printf("[MODEL] FOUND EXISTING COMPANY(%d)", c.ID)
	} else {
		// Before creating new Company, test if can just update the
		// calling object Company if only has one Security (assumed to
		// be the Security being updated). Racy! Maybe need the Save()
		// to be in a hook attached to the Find().
		if c.ID > 0 && tryUpdate {
			var numSecurities int64

			db.Model(&Security{}).Where(&Security{CompanyID: c.ID}).
			   Count(&numSecurities)
			if numSecurities == 1 {
				db.Save(c)
				log.Printf("[MODEL] UPDATED COMPANY(%d)", c.ID)
				return c
			}
		}

		db.Create(c)
		spewModel(c)
		log.Printf("[MODEL] CREATE COMPANY(%d)", c.ID)
	}

	return c
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

/* Return true if Symbol was updated. */
func (c *Company) update(db *gorm.DB) bool {
	if c.oldSymbol == c.Symbol {
		return false
	}

	newCompany := c.Get(db, true)
	if newCompany == nil {
		return false
	}
	c.ID = newCompany.ID
	return true
}
