/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package config

import (
	"log"
	"github.com/ilyakaznacheev/cleanenv"
)

var globalConfig *GlobalConfiguration

type GlobalConfiguration struct {
	Sessions bool `toml:"sessions"`
}

type Configuration struct {
	GlobalConfiguration GlobalConfiguration `toml:"global"`
}

func GlobalConfig() *GlobalConfiguration {
	return globalConfig
}

func getConfig() *Configuration {
	c := Configuration{}
	err := cleanenv.ReadConfig("config/database.toml", &c)
	if err != nil {
		log.Panic(err)
	}
	return &c
}

func init() {
	globalConfig = &getConfig().GlobalConfiguration
}
