// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"

	"github.com/galeone/sleepbit/client"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
)

func main() {
	hosts := map[string]*echo.Echo{}
	client, err := client.NewRouter()
	if err != nil {
		panic(err.Error())
	}

	port := os.Getenv("PORT")
	if port == "80" {
		port = ""
	} else {
		port = fmt.Sprintf(":%s", port)
	}
	hosts[fmt.Sprintf("client.%s%s", os.Getenv("DOMAIN"), port)] = client

	// Catch-all server & dispatch
	e := echo.New()
	e.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		host := hosts[req.Host]

		if host == nil {
			return echo.ErrNotFound
		}
		host.ServeHTTP(res, req)
		return
	})

	if port == "" {
		port = ":80"
	}
	e.Logger.Fatal(e.Start(port))
}
