/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package db

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"database/sql"
	"github.com/rubenv/sql-migrate"
)

//go:embed migrations/*.sql
var migrationDir embed.FS

func sqlMigrate(db *sql.DB, name string) {
	var err error
	n := 0
	useFS := true

	if useFS {
		httpDir,_ := fs.Sub(migrationDir, "migrations")
		migrations := &migrate.HttpFileSystemMigrationSource {
		    FileSystem: http.FS(httpDir),
		}
		n, err = migrate.Exec(db, name, migrations, migrate.Up)
	} else {
		migrations := &migrate.FileMigrationSource{
		    Dir: "db/migrations",
		}
		n, err = migrate.Exec(db, name, migrations, migrate.Up)
	}

	if err != nil {
		log.Panic(err)
	}
	log.Printf("[DB] MIGRATIONS APPLIED(%d)", n)
}
