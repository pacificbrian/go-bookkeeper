/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package route

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mayowa/echo-pongo2"
	"go-bookkeeper/controllers"
)

func UseTemplates(e *echo.Echo) echo.Renderer {
	views := NewTemplate()
	//views.Add("index.html", "public/views/base.html")

	return views
}

func UsePongo2(e *echo.Echo) echo.Renderer {
	r, err := echopongo2.NewRenderer("public/views/")
	if err != nil {
		r = nil
	}
	return r
}

func Init() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())    // log request info
	e.Use(middleware.Recover())   // auto recover from any panic
	e.Use(middleware.RequestID()) // log request info with id
	e.Static("/", "public")
	e.Renderer = UsePongo2(e)

	e.GET("/accounts", controllers.ListAccounts)   // Index/List
	e.POST("/accounts", controllers.CreateAccount) // Create
	//e.GET("/accounts/new", controllers.NewAccount)

	//e.GET("/accounts/:id/edit", controllers.EditAccount)
	e.GET("/accounts/:id", controllers.GetAccount) // Show
	e.POST("/accounts/:id", controllers.UpdateAccount) // Update
	e.DELETE("/accounts/:id", controllers.DeleteAccount)

	return e
}
