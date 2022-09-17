/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package db

import (
	"log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"go-bookkeeper/db/sqlite"
)

var db *gorm.DB

func init() {
	var err error

	config := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	db, err = sqlite.OpenSqlite(config)
	if err != nil {
		log.Panic(err)
	}

	autoMigrate(db)
}

func DbManager() *gorm.DB {
	return db
}

func DebugDbManager() *gorm.DB {
	return db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Info)})
}
