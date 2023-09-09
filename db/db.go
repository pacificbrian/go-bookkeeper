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

func Init() {
	var err error
	var name string

	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	driver := config.GetConfig().DB.DB
	switch driver {
	case "sqlite":
		name = sqlite.Name()
		db, err = gorm.Open(sqlite.Open(), gormConfig)
	case "mysql":
		name = mysql.Name()
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

	sqlMigrate(sqldb, name)
}

func DbManager() *gorm.DB {
	return db
}

func DebugDbManager() *gorm.DB {
	return db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Info)})
}
