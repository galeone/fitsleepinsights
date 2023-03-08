// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package client handles the client subdomain.
// The client subdomain is the subdomain that interacts
// with the Fitbit API

package app

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	_ "github.com/joho/godotenv/autoload"
)

func NewRouter() (*echo.Echo, error) {
	router := echo.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	router.Logger.SetLevel(log.DEBUG)

	router.GET("/auth", Auth())
	router.GET("/redirect", Redirect(), RequireFitbit())

	// Internal routes, used for:
	// Generating types and testing the Authorizer
	router.GET("/generate", GenerateTypes(), RequireFitbit())
	// Testing the API(*Authorizer)
	router.GET("/test", TestGET(), RequireFitbit())

	router.GET("/vertex", TestVertexAI())
	return router, nil
}
