/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package route

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Init() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())    // log request info
	e.Use(middleware.Recover())   // auto recover from any panic
	e.Use(middleware.RequestID()) // log request info with id
	e.Static("/", "public")

	views := NewTemplate()
	e.Renderer = views

	return e
}
