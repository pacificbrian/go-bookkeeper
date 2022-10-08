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

func (*CategoryType) List(db *gorm.DB) []CategoryType {
	entries := []CategoryType{}
	db.Find(&entries)

	return entries
}
