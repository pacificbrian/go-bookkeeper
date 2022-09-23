/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type Category struct {
	Model
	UserID uint
	User User
	CategoryTypeID uint `form:"category.category_type_id"`
	CategoryType CategoryType
	Name string `form:"category.Name"`
	OmitFromPie bool
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
