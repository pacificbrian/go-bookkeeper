# GO bookkeeper

A simple bookkeeping application written in Go. Server is using echo
for the router and gorm for ORM.  Client interface is using javascript
(stimulus) and HTML templates (pongo2).
  
## Overview

This is something I made for fun to learn Go, and to refresh my old
bookkeeping application that was using Rails.<br>
  
It's not ready to use yet, as is missing:
1. Delete support (was easy in Rails)
1. TradeCashFlows#Update

## Setup

```bash
make deps
make client
```

## Run

`go run server.go`

