/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package config

import (
	"log"
	"os"
	"github.com/ilyakaznacheev/cleanenv"
)

const moduleName string = "/github.pacificbrian.go-bookkeeper"
const configurationFile string = "/config.toml"
var configurationPath string

type Configuration struct {
	DB struct {
		DB       string `toml:"db" env:"GOBOOK_DB" env-default:"sqlite"`
		User     string `toml:"user" env:"GOBOOK_DB_USER" env-default:"root"`
		Password string `toml:"password" env:"GOBOOK_DB_PASSWORD"`
		Port     int    `toml:"port" env:"GOBOOK_DB_PORT" env-default:"3307"`
		Host     string `toml:"host" env:"GOBOOK_DB_HOST" env-default:"localhost"`
		Name     string `toml:"name" env:"GOBOOK_DB_NAME"`
	} `toml:"db"`
	GlobalConfiguration GlobalConfiguration `toml:"global"`
}

func GetConfigDir(localDir string) string {
	if DebugFlag {
		return localDir
	}
	var err error

	if configurationPath != "" {
		return configurationPath
	}

	homeDir := os.Getenv("GOBOOK_HOME")
	if homeDir == "" {
		homeDir, err = os.UserConfigDir()
		if err == nil {
			homeDir += moduleName
			err = os.Mkdir(homeDir, 0700)
		}
		if err != nil && !os.IsExist(err) {
			log.Panic(err)
		}

		file, err := os.OpenFile(homeDir + configurationFile,
					 os.O_RDONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Panic(err)
		}
		file.Close()
	}

	configurationPath = homeDir
	log.Printf("[CONFIG] USING DIRECTORY(%s/)", homeDir)
	return homeDir
}

func GetConfig() *Configuration {
	c := Configuration{}
	homeDir := GetConfigDir("config")

	err := cleanenv.ReadConfig(homeDir + configurationFile, &c)
	if err != nil {
		log.Panic(err)
	}
	return &c
}
