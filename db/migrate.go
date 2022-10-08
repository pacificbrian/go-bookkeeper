/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package db

import (
	"log"
	"database/sql"
	"gorm.io/gorm"
	"github.com/rubenv/sql-migrate"
	"go-bookkeeper/model"
)

func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(&model.AccountType{})
	db.AutoMigrate(&model.CategoryType{})
	db.AutoMigrate(&model.CashFlowType{})
	db.AutoMigrate(&model.Payee{})
	db.AutoMigrate(&model.Category{})
	db.AutoMigrate(&model.TradeType{})
	//db.AutoMigrate(&model.RepeatIntervalType{})
	//db.AutoMigrate(&model.RepeatInterval{})
	//db.AutoMigrate(&model.CurrencyType{})
	//db.AutoMigrate(&model.User{})
	//db.Debug().AutoMigrate(&model.Account{})
	//db.Debug().AutoMigrate(&model.CashFlow{})
}


func sqlMigrate(db *sql.DB, name string) {
	var migrations *migrate.FileMigrationSource
	use_packr := false

	if use_packr {
		//migrations = &migrate.PackrMigrationSource{
		//    Box: packr.New("migrations", "./migrations"),
		//}
	} else {
		migrations = &migrate.FileMigrationSource{
		    Dir: "db/migrations",
		}
	}

	n, err := migrate.Exec(db, name, migrations, migrate.Up)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("MIGRATIONS APPLIED(%d)", n)
}
