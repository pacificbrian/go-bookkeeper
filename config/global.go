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
	CashFlowLimit int `toml:"cashflow_limit"`
	EnableAutoTaxes bool `toml:"enable_auto_taxes" env-default:"true"`
	EnableImportTradeFixups bool `toml:"enable_import_trade_fixups" env-default:"false"`
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
