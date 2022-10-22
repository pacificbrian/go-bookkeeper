/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package main

import (
	"github.com/pacificbrian/go-bookkeeper/route"
)

func main() {
	e := route.Init()
	e.Logger.Fatal(e.Start(":3000"))
}
