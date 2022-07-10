/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package sql

import (
	"log"
	"github.com/ilyakaznacheev/cleanenv"
)

type Configuration struct {
	DB struct {
		User     string `toml:"user" env:"DB_USER" env-default:"root"`
		Password string `toml:"password" env:"DB_PASSWORD"`
		Port     int    `toml:"port" env:"DB_PORT" env-default:"3307"`
		Host     string `toml:"host" env:"DB_HOST" env-default:"localhost"`
		Name     string `toml:"name" env:"DB_NAME" env-default:"gobook_production"`
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
