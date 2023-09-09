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

func sqlMigrate(db *sql.DB, name string, direction migrate.MigrationDirection) int {
	var err error
	n := 0
	useFS := true

	if useFS {
		httpDir,_ := fs.Sub(migrationDir, "migrations")
		migrations := &migrate.HttpFileSystemMigrationSource {
		    FileSystem: http.FS(httpDir),
		}
		n, err = migrate.Exec(db, name, migrations, direction)
	} else {
		migrations := &migrate.FileMigrationSource{
		    Dir: "db/migrations",
		}
		n, err = migrate.Exec(db, name, migrations, direction)
	}

	if err != nil {
		log.Panic(err)
	}

	return n
}

func sqlMigrateUp(db *sql.DB, name string) {
	n := sqlMigrate(db, name, migrate.Up)
	log.Printf("[DB] APPLIED MIGRATIONS(%d)", n)
}

func sqlMigrateDown(db *sql.DB, name string) {
	n := sqlMigrate(db, name, migrate.Down)
	log.Printf("[DB] REVERSED MIGRATIONS(%d)", n)
}
