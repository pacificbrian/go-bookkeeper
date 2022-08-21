/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package db

import (
	"gorm.io/gorm"
	"go-bookkeeper/model"
)

func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(&model.AccountType{})
	//db.AutoMigrate(&model.CurrencyType{})
	//db.Debug().AutoMigrate(&model.Account{})
}
