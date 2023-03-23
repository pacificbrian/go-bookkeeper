/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package main

import (
	"context"
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/labstack/echo/v4"
	"github.com/pacificbrian/go-bookkeeper/controllers"
	"github.com/pacificbrian/go-bookkeeper/route"
)

//go:embed views
var templateStore embed.FS

func userSignal(ctx context.Context, e *echo.Echo) {
	select {
	case <-ctx.Done():
		log.Printf("[SERVER] CAUGHT SIGTERM")
		controllers.CloseActiveSessions()
		e.Shutdown(ctx)
	}
}

func main() {
	e := route.Init(&templateStore)

	ctx, stop := signal.NotifyContext(context.Background(),
					  os.Interrupt, syscall.SIGTERM)
	defer stop()
	go userSignal(ctx, e)

	err := e.Start(":3000")
	log.Printf("[SERVER] EXIT: %v", err)
}
