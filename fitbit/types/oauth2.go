// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import "time"

// AuthorizedUser represents the payload received
// after succesfully exchaing the Authorization Code
// with the Fitbit OAuth2 server (Server Application Type)
// See the [documentation] - step 4.
//
// [documentation]: https://dev.fitbit.com/build/reference/web-api/developer-guide/authorization/#Authorization-Code-Grant-Flow-with-PKCE
type AuthorizedUser struct {
	AccessToken  string    `json:"access_token"`
	ExpiresIn    int64     `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	TokenType    string    `json:"token_type"`
	UserID       string    `json:"user_id" igor:"primary_key"`
	CreatedAt    time.Time `sql:"default:now()"`
}

func (AuthorizedUser) TableName() string {
	return "oauth2_authorized"
}

// OAuth2ErrorMessage represents the basic unit used by the Fitbit API for
// sending messages during the OAuth2 authorization flow.
type OAuth2ErrorMessage struct {
	ErrorType string `json:"errorType"`
	Message   string `json:"message"`
}

// OAuth2Error represents the payload received in case of error, during the
// OAuth2 authorization flow
type OAuth2Error struct {
	Errors  []OAuth2ErrorMessage
	Success bool `json:"success"`
}

// AuthorizingUser is the type used during the exchange of the
// "Code" for the tokens in the OAuth2 flow.
// The CSRFToken is used as primary key, other than a CSRF token.
type AuthorizingUser struct {
	CSRFToken string `igor:"primary_key"`
	Code      string
	CreatedAt time.Time `sql:"default:now()"`
}

func (AuthorizingUser) TableName() string {
	return "oauth2_authorizing"
}
