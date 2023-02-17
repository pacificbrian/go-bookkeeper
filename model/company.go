/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
)

type Company struct {
	Model
	Name string `form:"Name"`
	Symbol string `form:"Symbol"`
	oldName string `gorm:"-:all"`
	oldSymbol string `gorm:"-:all"`
	// UserID only used from Security.Update
	UserID uint `gorm:"-:all"`
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
func (c *Company) Get(tryUpdate bool) *Company {
	if c.Name == "" && c.Symbol == "" {
		return nil
	}
	result := new(Company)
	db := getDbManager()

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

func (c *Company) updateAllowed() bool {
	return (c.Name != "" || c.Symbol != "")
}

// Updates Company Name, if Symbol unchanged.
// But do not modify a Company that is attached to the
// Securities of other Users.
// Return true if Company.Name was updated.
func (c *Company) updateName() bool {
	var err error

	if !c.updateAllowed() {
		return false
	}

	if c.oldSymbol == c.Symbol &&
	   c.oldName != c.Name {
		db := getDbManager()
		var numSecurities int64

		// Verify Company not used by other Users' Securities
		if c.UserID > 0 {
			db.Model(&Security{}).
			   Where(&Security{CompanyID: c.ID}).
			   Where("user_id != ?", c.UserID).
			   Joins("Account").Count(&numSecurities)
		}

		if numSecurities == 0 {
			result := db.Model(c).Update("Name", c.Name)
			err = result.Error
			if err == nil {
				log.Printf("[MODEL] UPDATED COMPANY(%d) NAME(%s)",
					   c.ID, c.Name)
				return true
			}
		}
	}
	return false
}

// Return true if Company.ID was updated.
func (c *Company) Update() bool {
	if !c.updateAllowed() ||
	   (c.oldSymbol == c.Symbol && c.oldName == c.Name) {
		return false
	}

	// First try update of just c.Name
	if c.updateName() {
		return false
	}

	newCompany := c.Get(true)
	if newCompany == nil {
		return false
	}
	c.ID = newCompany.ID
	return true
}
