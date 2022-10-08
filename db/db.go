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
	"gorm.io/gorm/logger"
	"go-bookkeeper/db/sqlite"
)

var sqldb *sql.DB
var db *gorm.DB

func init() {
	var err error

	config := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	name := sqlite.Name()
	db, err = gorm.Open(sqlite.Open(), config)
	if err == nil {
		sqldb, err = db.DB()
	}
	if err != nil {
		log.Panic(err)
	}

	sqlMigrate(sqldb, name)
}

func DbManager() *gorm.DB {
	return db
}

func DebugDbManager() *gorm.DB {
	return db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Info)})
}
