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
	entries := []Category{}
	db.Find(&entries)
	return entries
}
