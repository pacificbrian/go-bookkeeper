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
)

func Open() gorm.Dialector {
	c := getConfig()
	log.Printf("OPEN DATABASE(%s)", c.DB.Name)
	return sqlite.Open(c.DB.Name)
}

func Name() string {
	return "sqlite3"
}
