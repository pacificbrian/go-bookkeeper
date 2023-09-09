/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package db

import (
	"errors"
	"fmt"
	"log"
	"database/sql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/pacificbrian/go-bookkeeper/config"
	"github.com/pacificbrian/go-bookkeeper/db/mysql"
	"github.com/pacificbrian/go-bookkeeper/db/sqlite"
)

var sqldb *sql.DB
var db *gorm.DB
var dbType string

func Init() {
	allowReset := false
	var err error

	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	driver := config.GetConfig().DB.DB
	switch driver {
	case "sqlite":
		dbType = sqlite.Name()
		db, err = gorm.Open(sqlite.Open(), gormConfig)
		allowReset = sqlite.IsDefaultDB
	case "mysql":
		dbType = mysql.Name()
		db, err = gorm.Open(mysql.Open(), gormConfig)
	default:
		err = errors.New(fmt.Sprintf("Unknown Database choice (%s)!", driver))
	}

	if err == nil {
		sqldb, err = db.DB()
	}
	if err != nil {
		log.Panic(err)
	}

	sqlMigrateUp(sqldb, dbType)
	if !allowReset {
		dbType = ""
	}
}

func Reset() {
	if dbType != "" {
		log.Printf("[DB] RESET (%s)", dbType)
		sqlMigrateDown(sqldb, dbType)
	}
}

func DbManager() *gorm.DB {
	return db
}

func DebugDbManager() *gorm.DB {
	return db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Info)})
}
