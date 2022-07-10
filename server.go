/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package main

import (
	"go-bookkeeper/db"
	"go-bookkeeper/route"
)

func main() {
	db.Init()
	e := route.Init()
	e.Logger.Fatal(e.Start(":3000"))
}
