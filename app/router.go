// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package client handles the client subdomain.
// The client subdomain is the subdomain that interacts
// with the Fitbit API

package app

import (
	"html/template"
	"strings"
	"time"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/echoview-v4"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
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

	// use echoview(goview) as simplified renderer
	viewConf := goview.DefaultConfig
	viewConf.Funcs["year"] = time.Now().UTC().Year
	viewConf.Funcs["contains"] = strings.Contains
	viewConf.Funcs["hasPrefix"] = strings.HasPrefix
	viewConf.Funcs["hasSuffix"] = strings.HasSuffix
	viewConf.Funcs["nl2br"] = func(text string) template.HTML {
		return template.HTML(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1))
	}
	viewConf.Funcs["md2html"] = func(md string) template.HTML {
		// create markdown parser with extensions
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
		p := parser.NewWithExtensions(extensions)
		doc := p.Parse([]byte(md))

		// create HTML renderer with extensions
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)

		return template.HTML(markdown.Render(doc, renderer))
	}

	router.Renderer = echoview.New(viewConf)

	router.GET("/auth", Auth())
	router.GET("/redirect", Redirect())
	router.GET("/dashboard", Dashboard(), RequireFitbit())

	router.Static("/static", "static")

	// INTERNAL routes used for:

	// Generating types and testing the Authorizer
	router.GET("/generate", GenerateTypes(), RequireFitbit())
	// Testing the API(*Authorizer)
	router.GET("/test", TestGET(), RequireFitbit())
	// Dump all data endpoint (INTERNAL)
	router.GET("/dump", Dump(), RequireFitbit())
	// Train and deploy sleep efficiency predictor
	router.GET("/train", TestTrainAndDeploy(), RequireFitbit())
	// Predict sleep efficiency
	router.GET("/predict", TestPredictSleepEfficiency(), RequireFitbit())
	// Fetch all data endpoint (INTERNAL)
	router.GET("/fetch", Fetch(), RequireFitbit())
	return router, nil
}
