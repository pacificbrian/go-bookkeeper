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
	db.AutoMigrate(&model.CategoryType{})
	db.AutoMigrate(&model.CashFlowType{})
	db.AutoMigrate(&model.Payee{})
	db.AutoMigrate(&model.Category{})
	//db.AutoMigrate(&model.RepeatIntervalType{})
	//db.AutoMigrate(&model.RepeatInterval{})
	//db.AutoMigrate(&model.CurrencyType{})
	//db.AutoMigrate(&model.User{})
	//db.Debug().AutoMigrate(&model.Account{})
	//db.Debug().AutoMigrate(&model.CashFlow{})
}
