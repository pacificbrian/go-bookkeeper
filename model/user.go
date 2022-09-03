/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Login string
	Email string
	Password string `gorm:"->:false;<-"`
	Categories []Category
	Payees []Payee
}
