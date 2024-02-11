// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package client handles the client subdomain.
// The client subdomain is the endpoint for commnuicating with
// the Fitbit API.

package app

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/galeone/fitbit/v2"
	"github.com/galeone/fitbit/v2/types"
	"github.com/galeone/fitsleepinsights/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Auth redirects the user to the fitbit authorization page
// It sets a cookie the univocally identifies the user
// because the authorizer.Exchange (used in Redirect)
// needs to check the `code` and CSRF tokens - and these tokens
// are attributes of the fitbit client that needs to persist
// from Auth() to Redirect().
// NOTE: it uses the environment variables:
// - FITBIT_CLIENT_ID
// - FITBIT_CLIENT_SECRET
// - FITBIT_REDIRECT_URL
// Loaded from a .env file - if any.
func Auth() func(echo.Context) error {
	return func(c echo.Context) (err error) {
		authorizer := fitbit.NewAuthorizer(_db, _clientID, _clientSecret, _redirectURL)

		authorizing := types.AuthorizingUser{
			CSRFToken: uuid.New().String(),
			// Code verifier for PKCE
			// https://dev.fitbit.com/build/reference/web-api/developer-guide/authorization/#Authorization-Code-Grant-Flow-with-PKCE
			Code: fmt.Sprintf("%s-%s", uuid.New().String(), uuid.New().String()),
		}

		authorizer.SetAuthorizing(&authorizing)

		c.SetCookie(&http.Cookie{
			Name: "authorizing",
			// Also used as primary key in db for retrieval (see middleware
			// RequireAuthorizer).
			Value: authorizer.CSRFToken().String(),
			// No Expires = Session cookie
			HttpOnly: true,
			Path:     "/",
		})

		// Every time we are in /auth, we want to remove the token cookie
		if _, err = c.Cookie("token"); err == nil {
			c.SetCookie(&http.Cookie{
				Name:   "token",
				MaxAge: -1,
				Path:   "/",
			})
		}

		if err = _db.InsertAuthorizingUser(&authorizing); err != nil {
			return err
		}

		var auth_url *url.URL
		if auth_url, err = authorizer.AuthorizationURL(); err != nil {
			return err
		}

		return c.Redirect(http.StatusTemporaryRedirect, auth_url.String())
	}
}

// Redirect handles the redirect from the Fitbit API to our redirect URI.
// Sets the "token" cookie for the whole domain, containing the access token
// The access token univocally identifies the user. The token expires when the
// access token expires.
func Redirect() func(echo.Context) error {
	return func(c echo.Context) (err error) {
		authorizer := fitbit.NewAuthorizer(_db, _clientID, _clientSecret, _redirectURL)
		var cookie *http.Cookie
		if cookie, err = c.Cookie("authorizing"); err == nil {
			var authorizing *types.AuthorizingUser
			if authorizing, err = _db.AuthorizingUser(cookie.Value); err != nil {
				log.Printf("[RequireFitbit] _db.AuthorizingUser: %s", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/auth")
			}
			authorizer.SetAuthorizing(authorizing)
		} else {
			// No cookies set
			return c.Redirect(http.StatusTemporaryRedirect, "/auth")
		}

		state := c.QueryParam("state")
		if state != authorizer.CSRFToken().String() {
			c.Logger().Warnf("Invalid state in /redirect. Expected %s got %s", authorizer.CSRFToken().String(), state)
			return c.Redirect(http.StatusTemporaryRedirect, "/error?status=csrf")
		}

		code := c.QueryParam("code")
		var token *types.AuthorizedUser
		if token, err = authorizer.ExchangeAuthorizationCode(code); err != nil {
			c.Logger().Warnf("ExchangeAuthorizationCode: %s", err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, "/error?status=exchange")
		}
		// Update the fitbitclient. Now it contains a valid token and HTTP can be used to query the API
		authorizer.SetToken(token)

		// Save token and redirect user to the application dashboard
		if err = _db.UpsertAuthorizedUser(token); err != nil {
			return err
		}
		// Send a database notification over the channel.
		// The receiver will start the routing for fetching all the data
		if err = _db.Notify(database.NewUsersChannel, token.AccessToken); err != nil {
			c.Logger().Error("Unable to sent new user creation notification")
		}
		cookie = &http.Cookie{
			Name:     "token",
			Value:    token.AccessToken,
			Domain:   _domain,
			Expires:  time.Now().Add(time.Second * time.Duration(token.ExpiresIn)),
			HttpOnly: true,
		}
		c.SetCookie(cookie)

		// Unset the authorizing cookie
		c.SetCookie(&http.Cookie{
			Name:     "authorizing",
			HttpOnly: true,
			MaxAge:   -1,
			Path:     "/",
		})
		return c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
	}
}

func Error() func(echo.Context) error {
	return func(c echo.Context) error {
		status := c.QueryParam("status")

		type ErrorMessage struct {
			Error string `json:"error"`
		}
		switch status {
		case "csrf":
			return c.JSON(http.StatusBadRequest, ErrorMessage{
				Error: "CSRF attempt detected",
			})
		case "exchange":
			return c.JSON(http.StatusBadGateway, ErrorMessage{
				Error: "Error exchanging OAuth2 Authorization code for the tokens",
			})
		}
		return nil
	}
}
