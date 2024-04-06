package app

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func Privacy() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		return c.Render(http.StatusOK, "privacy", echo.Map{
			"title":      "Privacy Policy - FitSleepInsights",
			"isLoggedIn": validLogin(c),
			"domain":     os.Getenv("DOMAIN"),
		})
	}
}

func Contact() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		return c.Render(http.StatusOK, "contact", echo.Map{
			"title":      "Contact - FitSleepInsights",
			"isLoggedIn": validLogin(c),
			"domain":     os.Getenv("DOMAIN"),
		})
	}
}

func About() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		return c.Render(http.StatusOK, "about", echo.Map{
			"title":      "About - FitSleepInsights",
			"isLoggedIn": validLogin(c),
			"domain":     os.Getenv("DOMAIN"),
		})
	}
}

func Index() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if validLogin(c) {
			return c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
		}
		return c.Render(http.StatusOK, "index", echo.Map{
			"title":      "FitSleepInsights",
			"isLoggedIn": false,
			"domain":     os.Getenv("DOMAIN"),
		})
	}
}
