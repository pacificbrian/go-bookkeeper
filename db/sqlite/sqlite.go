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

func OpenSqlite(gconfig *gorm.Config) (*gorm.DB, error) {
	c := getConfig()
	log.Printf("OPEN DATABASE(%s)", c.DB.Name)
	sqldb := sqlite.Open(c.DB.Name)
	return gorm.Open(sqldb, gconfig)
}
