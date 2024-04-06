// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"log"
	"net/http"

	"github.com/galeone/fitbit/v2"
	"github.com/galeone/fitbit/v2/types"
	"github.com/labstack/echo/v4"
)

// RequireFitbit is the middleware to use when a route requires
// to interact with the fitbit API.
// The middleware uses the cookies to identify the user and
// understand in which phase of the oauth2 authorization flows we are
// and set the context's fitbit variable (c.Get("fitbit")) to a valid authorizer
// If and only if the required cookies have been previously set.
func RequireFitbit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if c.Get("fitbit") == nil {
				// The authorizing cookie handling is the same that we do in /redirect
				// (see routes_oauth2.go).
				// We do it also here because RequireFitbit is the middleware that
				// is used in all the routes that require the fitbit API.
				// And we check that if authorizing is set, than we should do the auth + redirect flow
				// for some reason (maybe the user has deleted the cookies).

				authorizer := fitbit.NewAuthorizer(_db, _clientID, _clientSecret, _redirectURL)
				var cookie *http.Cookie
				if cookie, err = c.Cookie("authorizing"); err == nil {
					var authorizing *types.AuthorizingUser
					if authorizing, err = _db.AuthorizingUser(cookie.Value); err != nil {
						log.Printf("[RequireFitbit] _db.AuthorizingUser: %s", err)
						return c.Redirect(http.StatusTemporaryRedirect, "/auth")
					}
					authorizer.SetAuthorizing(authorizing)
					// This route requires the token or the auth cookie
					return c.Redirect(http.StatusTemporaryRedirect, "/auth")
				}

				// Authorization token (after exchange)
				if cookie, err = c.Cookie("token"); err != nil {
					// No cookies set
					return c.Redirect(http.StatusTemporaryRedirect, "/auth")
				}

				var dbToken *types.AuthorizedUser
				if dbToken, err = _db.AuthorizedUser(cookie.Value); err != nil {
					log.Printf("[RequireFitbit] _db.AuthorizedUser: %s", err)
					return c.Redirect(http.StatusTemporaryRedirect, "/auth")
				}
				if dbToken.UserID == "" {
					log.Println(err)
					return c.Redirect(http.StatusTemporaryRedirect, "/auth")
				}
				authorizer.SetToken(dbToken)
				c.Set("fitbit", authorizer)
			}
			return next(c)
		}
	}
}

func validLogin(c echo.Context) bool {
	// Authorization token (after exchange)
	var cookie *http.Cookie
	var err error
	if cookie, err = c.Cookie("token"); err != nil {
		// No cookies set
		return false
	}

	var dbToken *types.AuthorizedUser
	if dbToken, err = _db.AuthorizedUser(cookie.Value); err != nil {
		return false
	}
	if dbToken.UserID == "" {
		return false
	}
	return true
}
