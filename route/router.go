/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package route

import (
	"embed"
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mayowa/echo-pongo2"
	"github.com/spazzymoto/echo-scs-session"
	"github.com/pacificbrian/go-bookkeeper/config"
	"github.com/pacificbrian/go-bookkeeper/controllers"
)

func usePongo2(e *echo.Echo, useEmbed *embed.FS) echo.Renderer {
	if config.DebugFlag {
		useEmbed = nil
	}
	opts := echopongo2.Options{UseEmbed: useEmbed,
				   Debug: config.DebugFlag}
	r, err := echopongo2.NewRenderer("views/", opts)
	if err != nil {
		r = nil
	}
	return r
}

func Init(staticFS *embed.FS, viewFS *embed.FS) *echo.Echo {
	sMan := controllers.StartSessionManager()
	e := echo.New()

	e.Use(middleware.Logger())    // log request info
	e.Use(middleware.Recover())   // auto recover from any panic
	e.Use(middleware.RequestID()) // log request info with id
	if sMan != nil {
		e.Use(session.LoadAndSave(sMan))
	}

	if config.DebugFlag {
		e.Static("/", "public")
	} else {
		staticConfig := middleware.StaticConfig {
			Root:       "public",
			Filesystem: http.FS(staticFS),
		}
		e.Use(middleware.StaticWithConfig(staticConfig))
	}

	e.Renderer = usePongo2(e, viewFS)

	// Login (or default)
	e.GET("/", controllers.Login)
	e.POST("/sessions", controllers.CreateSession)

	// User
	e.GET("/users/new", controllers.NewUser)
	e.POST("/users", controllers.CreateUser)

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
	e.GET("/accounts/:id/imported", controllers.ListImported)
	e.GET("/imported/:id", controllers.ListImportedCashFlows)

	// Payee
	e.GET("/payees", controllers.ListPayees)
	e.GET("/accounts/:account_id/payees", controllers.ListPayees)
	e.POST("/payees", controllers.CreatePayee)
	e.GET("/payees/:id/edit", controllers.EditPayee)
	e.GET("/payees/:id", controllers.GetPayee)
	e.GET("/accounts/:account_id/payees/:id", controllers.GetPayee)
	e.POST("/payees/:id", controllers.UpdatePayee)
	e.POST("/accounts/:account_id/payees/:id", controllers.UpdatePayee)
	e.POST("/payees/:id/merge", controllers.MergePayee)
	e.POST("/accounts/:account_id/payees/:id/merge", controllers.MergePayee)
	e.DELETE("/payees/:id", controllers.DeletePayee)

	// Company
	//e.GET("/companies", controllers.ListCompanies)
	e.GET("/companies/:company_id/financials", controllers.GetCompanyFinancials)

	// Security
	e.GET("/accounts/:account_id/securities/:id", controllers.GetSecurity) // Show
	e.GET("/accounts/:account_id/securities/new", controllers.NewSecurity)
	e.POST("/accounts/:account_id/securities", controllers.CreateSecurity)
	e.GET("/securities/:id/edit", controllers.EditSecurity)
	e.POST("/securities/:id", controllers.UpdateSecurity)
	e.GET("/securities/:id/move", controllers.MoveSecurity)
	e.POST("/securities/:id/move", controllers.MoveSecurity)

	// Trade
	e.POST("/accounts/:account_id/trades", controllers.CreateTrade)
	e.POST("/securities/:security_id/trades", controllers.CreateTrade)
	e.GET("/trades/:id/edit", controllers.EditTrade)
	e.POST("/securities/:security_id/trades/:id", controllers.UpdateTrade)
	e.POST("/trades/:id", controllers.UpdateTrade)
	e.DELETE("/trades/:id", controllers.DeleteTrade)
	e.GET("/years/:year/gains", controllers.ListTradeGains)
	e.GET("/years/:year/accounts/:account_id/gains", controllers.ListTradeGains)
	e.GET("/gains/:id", controllers.GetTradeGain)

	// Taxes
	e.GET("/years/:year/taxes", controllers.ListTaxes)
	e.GET("/taxes", controllers.ListTaxes)
	e.POST("/taxes", controllers.CreateTaxes)
	e.PUT("/taxes/:id", controllers.RecalculateTaxes)
	e.DELETE("/taxes/:id", controllers.DeleteTaxes)
	e.POST("/tax_entries", controllers.CreateTaxEntry)
	e.DELETE("/tax_entries/:id", controllers.DeleteTaxEntry)
	e.GET("/years/:year/tax_items/:tax_item_id", controllers.ListTaxCashFlows)
	e.GET("/years/:year/tax_types/:tax_type_id", controllers.ListTaxCashFlows)
	e.GET("/tax_entries/:id/edit", controllers.EditTaxEntry)
	e.POST("/tax_entries/:id", controllers.UpdateTaxEntry)
	//e.GET("/tax_categories", controllers.ListTaxCategories)
	//e.GET("/tax_years", controllers.ListTaxYears)

	return e
}
