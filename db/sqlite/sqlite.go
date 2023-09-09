/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package sqlite

import (
	"log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/pacificbrian/go-bookkeeper/config"
)

const defaultDatabaseName string = "/gobook_test.db"
var IsDefaultDB bool

func Open() gorm.Dialector {
	c := config.GetConfig()
	if c.DB.Name == "" {
		dir := config.GetConfigDir("db")
		c.DB.Name = dir + defaultDatabaseName
		IsDefaultDB = true
	}
	log.Printf("[DB] OPEN DATABASE(%s)", c.DB.Name)
	return sqlite.Open(c.DB.Name)
}

func Name() string {
	return "sqlite3"
}
