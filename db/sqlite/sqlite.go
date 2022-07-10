/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package sqlite

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func OpenSqlite() (*gorm.DB, error) {
	c := getConfig()
	sqldb := sqlite.Open(c.DB.Name)
	return gorm.Open(sqldb, &gorm.Config{})
}
