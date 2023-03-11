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

const defaultDatabaseName string = "db/gobook_test.db"

func Open(debug bool) gorm.Dialector {
	c := config.GetConfig(debug)
	if c.DB.Name == "" {
		c.DB.Name = defaultDatabaseName
	}
	log.Printf("OPEN DATABASE(%s)", c.DB.Name)
	return sqlite.Open(c.DB.Name)
}

func Name() string {
	return "sqlite3"
}
