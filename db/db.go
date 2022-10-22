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
	"github.com/ilyakaznacheev/cleanenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/pacificbrian/go-bookkeeper/db/mysql"
	"github.com/pacificbrian/go-bookkeeper/db/sqlite"
)

type Configuration struct {
	DB struct {
		DB string `toml:"db" env:"GOBOOK_DB" env-default:"sqlite"`
	} `toml:"db"`
}

var sqldb *sql.DB
var db *gorm.DB

func getConfig() *Configuration {
	c := Configuration{}
	err := cleanenv.ReadConfig("config/database.toml", &c)
	if err != nil {
		log.Panic(err)
	}
	return &c
}

func init() {
	var err error
	var name string

	config := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	driver := getConfig().DB.DB
	switch driver {
	case "sqlite":
		name = sqlite.Name()
		db, err = gorm.Open(sqlite.Open(), config)
	case "mysql":
		name = mysql.Name()
		db, err = gorm.Open(mysql.Open(), config)
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
