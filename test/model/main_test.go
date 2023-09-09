/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model_test

import (
	"log"
	"os"
	"testing"
	"github.com/pacificbrian/go-bookkeeper/controllers"
	"github.com/pacificbrian/go-bookkeeper/db"
	"github.com/pacificbrian/go-bookkeeper/model"
)

var defaultSession *model.Session

func setupTestEnv() {
	log.Println("[TEST] SET ENVIRONMENT")
	err := os.Setenv("GOBOOK_HOME", ".")
	if err != nil {
		log.Fatal("Error setting GOBOOK_HOME")
	}
}

func TestMain(m *testing.M) {
	log.Println("[TEST] MODEL MAIN START")

	// below must be in this order
	setupTestEnv()
	db.Init()
	defaultSession = controllers.CreateDefaultSession()

	err := m.Run()

	db.Reset()
	log.Println("[TEST] MODEL MAIN EXIT")
	os.Exit(err)
}
