/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package db

import (
	"log"
	"gorm.io/gorm"
	"go-bookkeeper/db/sqlite"
)

var db *gorm.DB

func Init() {
	var err error

	db, err = sqlite.OpenSqlite()
	if err != nil {
		log.Panic(err)
	}
}

func DbManager() *gorm.DB {
	return db
}
