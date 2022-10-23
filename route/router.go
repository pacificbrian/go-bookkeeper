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
	"github.com/pacificbrian/go-bookkeeper/controllers"
)

func usePongo2(e *echo.Echo) echo.Renderer {
	r, err := echopongo2.NewRenderer("views/")
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
	e.Renderer = usePongo2(e)

	// Login (or default)
	e.GET("/", controllers.Login)

	// Account
	e.GET("/accounts", controllers.ListAccounts)   // Index/List
	e.POST("/accounts", controllers.CreateAccount) // Create
	e.GET("/accounts/new", controllers.NewAccount)
	e.GET("/accounts/:id/edit", controllers.EditAccount)
	e.GET("/accounts/:id", controllers.GetAccount) // Show
	e.POST("/accounts/:id", controllers.UpdateAccount) // Update
	e.DELETE("/accounts/:id", controllers.DeleteAccount)

	// CashFlow
	e.POST("/accounts/:id/cash_flows", controllers.CreateCashFlow)
	e.POST("/accounts/:id/scheduled", controllers.CreateScheduledCashFlow)
	e.GET("/accounts/:id/scheduled", controllers.ListScheduledCashFlows)
	e.POST("/cash_flows/:id/split", controllers.CreateSplitCashFlow)
	e.GET("/cash_flows/:id/edit", controllers.EditCashFlow)
	e.POST("/cash_flows/:id", controllers.UpdateCashFlow)
	e.PUT("/cash_flows/:id", controllers.PutCashFlow)
	e.DELETE("/cash_flows/:id", controllers.DeleteCashFlow)

	// Import
	e.POST("/accounts/:id/imported", controllers.CreateImportedCashFlows)
	e.GET("/accounts/:id/imported", controllers.ListImportedCashFlows)

	// Payee
	e.GET("/payees", controllers.ListPayees)
	e.POST("/payees", controllers.CreatePayee)
	//e.GET("/payees/:id/edit", controllers.EditPayee)
	//e.GET("/payees/:id", controllers.GetPayee)
	e.POST("/payees/:id", controllers.UpdatePayee)

	// Security
	e.GET("/accounts/:account_id/securities/:id", controllers.GetSecurity) // Show

	// Trade
	e.POST("/accounts/:account_id/trades", controllers.CreateTrade)
	e.POST("/securities/:security_id/trades", controllers.CreateTrade)

	// Taxes
	e.GET("/years/:year/taxes", controllers.ListTaxes)
	e.GET("/taxes", controllers.ListTaxes)
	e.POST("/taxes", controllers.CreateTaxes)
	e.PUT("/taxes/:id", controllers.RecalculateTaxes)
	e.DELETE("/taxes/:id", controllers.DeleteTaxes)
	e.POST("/tax_entries", controllers.CreateTaxEntry)
	//e.GET("/tax_entries/:id/edit", controllers.EditTaxEntry)
	//e.POST"/tax_entries/:id", controllers.UpdateTaxEntry)
	//e.GET("/tax_categories", controllers.ListTaxCategories)
	//e.GET("/tax_years", controllers.ListTaxYears)

	return e
}
