/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type RepeatIntervalType struct {
	Model
	Name string `form:"repeat_interval_type.Name"`
	Days uint
}

func (*RepeatIntervalType) List(db *gorm.DB) []RepeatIntervalType {
	entries := []RepeatIntervalType{}
	db.Find(&entries)

	return entries
}
