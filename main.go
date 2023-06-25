// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/galeone/fitsleepinsights/app"
	_ "github.com/galeone/fitsleepinsights/database"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
)

func main() {
	domains := map[string]*echo.Echo{}
	app, err := app.NewRouter()
	if err != nil {
		panic(err.Error())
	}

	port := os.Getenv("PORT")
	if port == "80" {
		port = ""
	} else {
		port = fmt.Sprintf(":%s", port)
	}
	log.Printf("Locally, you can visit: %s%s\n", os.Getenv("DOMAIN"), port)

	// Hosts without port, because reverse proxies do not forward the port
	domains[os.Getenv("DOMAIN")] = app

	// Catch-all server & dispatch
	e := echo.New()
	e.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		// remove eventual port from req.Host
		// so this mapping works also locally
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			// no port found, use the whole host
			host = req.Host
		}

		if server, ok := domains[host]; !ok {
			return echo.ErrNotFound
		} else {
			server.ServeHTTP(res, req)
		}
		return
	})

	if port == "" {
		port = ":80"
	}

	e.Logger.Fatal(e.Start(port))
}
