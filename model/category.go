/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type CategoryType struct {
	Model
	Name string `form:"category_type.Name"`
}

type Category struct {
	Model
	UserID uint
	User User
	CategoryTypeID uint `form:"category.category_type_id"`
	CategoryType CategoryType
	Name string `form:"category.Name"`
	OmitFromPie bool
}

func (c *Category) IsInterest() bool {
	return (c.ID == 34)
}

func (c *Category) IsMortgageInterest() bool {
	return (c.ID == 35)
}

func (c *Category) IsInterestIncome() bool {
	return (c.ID == 74)
}

func (c *Category) LoanPI() bool {
	return (c.IsInterest() || c.IsMortgageInterest())
}

func (*CategoryType) List(db *gorm.DB) []CategoryType {
	entries := []CategoryType{}
	db.Find(&entries)

	return entries
}

func (*Category) List(db *gorm.DB) []Category {
	// need userCache lookup
	var entries []Category
	sub_entries := []Category{}

	// Expenses
	db.Order("Name").Where("category_type_id < 2 OR category_type_id == 3").
			 Find(&sub_entries)
	entries = append(entries, sub_entries...)

	// Income
	db.Order("Name").Where("category_type_id == 0 OR category_type_id == 2").
			 Find(&sub_entries)
	entries = append(entries, sub_entries...)

	return entries
}
