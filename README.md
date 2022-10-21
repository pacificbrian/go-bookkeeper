# GO bookkeeper

A simple bookkeeping application written in Go. Server is using echo
for the router and gorm for ORM.  Client interface is using javascript
(stimulus) and HTML templates (pongo2).
  
## Overview

This is something I made for fun to learn Go, and to refresh my old
bookkeeping application that was using Rails.<br>
  
It's not yet complete enough to use as bookeeping solution yet, as is missing:
1. unit tests
1. validation of inputs
1. Trade#Edit/Update
1. can't Delete Payees, Trades, TaxItems
1. user login, session management

## Database

The database details should be specified in config/database.toml (see comment there).
Or by default, sqlite3 is used and database is created at: db/gobook_test.db.

## Setup / Install

Install Go and yarn (javascript package manager) using your favorite package manager.  You can run 'make first-time' to install with Brew (Mac OSX).

Then install other Go and javascript dependencies and Build/Install:
```bash
make first-time
make install
```

## Run

`~/go/bin/go-bookkeeper`

## Run (without installing)

```bash
make deps
go run server.go
```

