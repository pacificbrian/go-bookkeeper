/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package db

import (
	"log"
	"database/sql"
	"github.com/rubenv/sql-migrate"
)

func sqlMigrate(db *sql.DB, name string) {
	var migrations *migrate.FileMigrationSource
	use_packr := false

	if use_packr {
		//migrations = &migrate.PackrMigrationSource{
		//    Box: packr.New("migrations", "./migrations"),
		//}
	} else {
		migrations = &migrate.FileMigrationSource{
		    Dir: "db/migrations",
		}
	}

	n, err := migrate.Exec(db, name, migrations, migrate.Up)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("[DB] MIGRATIONS APPLIED(%d)", n)
}
