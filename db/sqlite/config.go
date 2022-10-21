/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package sqlite

import (
	"log"
	"github.com/ilyakaznacheev/cleanenv"
)

type Configuration struct {
	DB struct {
		Name     string `toml:"name" env:"GOBOOK_DB_NAME" env-default:"db/gobook_test.db"`
	} `toml:"db"`
}

func getConfig() *Configuration {
	c := Configuration{}
	err := cleanenv.ReadConfig("config/database.toml", &c)
	if err != nil {
		log.Panic(err)
	}
	return &c
}
