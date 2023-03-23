/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package mysql

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/pacificbrian/go-bookkeeper/config"
)

const defaultDatabaseName string = "gobook_production"

func Open() gorm.Dialector {
	c := config.GetConfig()
	if c.DB.Name == "" {
		c.DB.Name = defaultDatabaseName
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			   c.DB.User, c.DB.Password,
			   c.DB.Host, c.DB.Port, c.DB.Name)
	return mysql.Open(dsn)
}

func Name() string {
	return "mysql"
}
