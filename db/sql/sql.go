/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package sql

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func OpenMaria() (*gorm.DB, error) {
	c := getConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			   c.DB.User, c.DB.Password,
			   c.DB.Host, c.DB.Port, c.DB.Name)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
