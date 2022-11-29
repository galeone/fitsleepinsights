// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"errors"
	"net/http"

	"github.com/galeone/fitbit"
	"github.com/galeone/fitbit/types"
	"github.com/labstack/echo/v4"
)

// RequireFitbit is the middleware to use when a route requires
// to interact with the fitbit API.
// The middleware uses the cookies to identify the user and
// understand in which phase of the oauth2 authorization flows we are
// and set the context's fitbit variable (c.Get("fitbit")) to a valid authorizer
// If and only if the required cookies have been previosly set.
func RequireFitbit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if c.Get("fitbit") == nil {
				authorizer := fitbit.NewAuthorizer(_db, _clientID, _clientSecret, _redirectURL)

				// If there's the auth cookie, we could be in the
				// authorization phase, and thus we set it.
				// Anyway, if it's not present, it's not a problem IF
				// and only IF  there's the "token" cookie that contains
				// the access token.
				// At least one of these 2 conditions should be met
				var condition bool
				var cookie *http.Cookie
				if cookie, err = c.Cookie("authorizing"); err == nil {
					var authorizing *types.AuthorizingUser
					if authorizing, err = _db.AuthorizingUser(cookie.Value); err != nil {
						return err
					}
					authorizer.SetAuthorizing(authorizing)
					condition = true
				}

				// Auhtorization token (after exhange)
				if cookie, err = c.Cookie("token"); err == nil {
					var dbToken *types.AuthorizedUser
					if dbToken, err = _db.AuthorizedUser(cookie.Value); err != nil {
						return err
					}
					if dbToken.UserID == "" {
						return errors.New("Invalid token. Please login again")
					}
					authorizer.SetToken(dbToken)
					condition = true
				}

				if !condition {
					// This route requires the token or the auth cookie
					return c.Redirect(http.StatusTemporaryRedirect, "/auth")
				}

				c.Set("fitbit", authorizer)
			}
			return next(c)
		}
	}
}
