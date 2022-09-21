/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type SecurityType struct {
	Model
	Name string `form:"security_type.Name"`
}

func (*SecurityType) List(db *gorm.DB) []SecurityType {
	// need userCache lookup
	entries := []SecurityType{}
	db.Find(&entries)

	return entries
}
