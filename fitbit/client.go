// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package fitbit

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"

	"github.com/galeone/sleepbit/fitbit/scopes"
	"github.com/galeone/sleepbit/fitbit/types"
	"github.com/google/uuid"

	_ "github.com/joho/godotenv/autoload"
)

type FitbitClient struct {
	config      *oauth2.Config
	authorizing *types.AuthorizingUser
	token       *oauth2.Token
	userID      *string
	db          Storage
}

// NewClient creates a new FitbitClient. The created client
// is already configured for requesting the correct scopes and
// make authenticathed/authorized requests to the fitbit API.
// The db parameter must be a valid implementation of the Storage
// interface.
func NewClient(db Storage) *FitbitClient {
	client := FitbitClient{}

	client.db = db

	// OAuth2 Client configuration
	config := &oauth2.Config{}
	config.ClientID = os.Getenv("FITBIT_CLIENT_ID")
	config.ClientSecret = os.Getenv("FITBIT_CLIENT_SECRET")
	config.RedirectURL = os.Getenv("FITBIT_REDIRECT_URL")
	config.Endpoint = fitbit.Endpoint
	config.Scopes = scopes.All()
	client.config = config

	return &client
}

// SetAuthorizing sets the parameters required during the autorization process
func (c *FitbitClient) SetAuthorizing(auth *types.AuthorizingUser) {
	c.authorizing = auth
}

// AuthorizationURL returns the URL where to redirect the user
// where it will be asked for giving the permissions for the various scopes
func (c *FitbitClient) AuthorizationURL() (*url.URL, error) {

	if c.authorizing == nil {
		return nil, errors.New("AuthorizationURL called without setting Authorizing parameters first")
	}
	// The OAuth2 library creates an url with:
	// - scopes
	// - access_type
	// - client_id
	// - redirect_uri
	// - response_type=code
	// - state=`c.authorizing.CSRFToken`
	ret, _ := url.Parse(c.config.AuthCodeURL(c.authorizing.CSRFToken, oauth2.AccessTypeOffline))

	// But the Fitbit API also requires
	// https://dev.fitbit.com/build/reference/web-api/developer-guide/authorization/#Authorization-Code-Grant-Flow-with-PKCE
	// - code_challenge
	// - code_challenge_method=S256
	values := ret.Query()
	values.Add("code_challenge_method", "S256")

	// base64UrlEncode(sha256Hash(code_verifier))
	h := sha256.New()
	h.Write([]byte(c.authorizing.Code))
	shaSum := h.Sum(nil)
	challenge := base64.RawURLEncoding.EncodeToString(shaSum)
	values.Add("code_challenge", challenge)
	ret.RawQuery = values.Encode()
	return ret, nil
}

// CSRFToken returns the CSRF code associated with this session
func (c *FitbitClient) CSRFToken() *uuid.UUID {
	token := uuid.MustParse(c.authorizing.CSRFToken)
	return &token
}

// ExchangeAuthorizationCode exchanges the authorization code for the access
// and refresh tokens.
// In a Server Application Type, this request should be authenticated
// https://dev.fitbit.com/build/reference/web-api/developer-guide/authorization/#Authorization-Code-Grant-Flow-with-PKCE
// See step 4
//
// This method also saves the exchanged token inside the *FitbitClient structure. This token
// is later used for creating the HTTP client (see HTTP method).
func (c *FitbitClient) ExchangeAuthorizationCode(code string) (token *types.AuthorizedUser, err error) {
	// Manually build everyting because adding a custom header in the code exchange request is not supported
	// The URL creation is kept from there
	// https://github.com/golang/oauth2/blob/2e8d9340160224d36fd555eaf8837240a7e239a7/oauth2.go#L213

	v := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.config.RedirectURL},

		// From step 4
		"client_id":     {os.Getenv("FITBIT_CLIENT_ID")},
		"code_verifier": {c.authorizing.Code},
	}

	endpoint, _ := url.Parse(fmt.Sprintf("%s?%s", fitbit.Endpoint.TokenURL, v.Encode()))
	var req *http.Request
	if req, err = http.NewRequest("POST", endpoint.String(), nil); err != nil {
		return nil, err
	}

	auth := base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", os.Getenv("FITBIT_CLIENT_ID"), os.Getenv("FITBIT_CLIENT_SECRET"))))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))

	client := http.Client{}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(res.Body)

	expected := types.AuthorizedUser{}
	if err = json.Unmarshal(body, &expected); err != nil {
		unexpected := types.OAuth2Error{}
		if err = json.Unmarshal(body, &unexpected); err == nil {
			var sb strings.Builder
			last := len(unexpected.Errors) - 1
			for i, err := range unexpected.Errors {
				sb.WriteString(err.Message)
				if i != last {
					sb.WriteRune(',')
				}
			}
			return nil, errors.New(sb.String())
		}
		return nil, fmt.Errorf("Unexpected return body: %s", string(body))
	}
	return &expected, nil
}

// Return an HTTP client that uses the specified token for authenticating
// It handles all the refresh-token stuff, and it updates inside the db
// The values for the user that's this *FitbitClient
func (c *FitbitClient) HTTP() (*http.Client, error) {
	tokenSource := c.config.TokenSource(context.Background(), c.token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	if c.token.AccessToken != newToken.AccessToken {
		var dbToken *types.AuthorizedUser
		if dbToken, err = c.db.AuthorizedUser(c.token.AccessToken); err != nil {
			return nil, err
		}

		// Now I have the dbToken that contains the UserID (primary key)
		// associated with the previous access token
		dbToken.AccessToken = newToken.AccessToken
		dbToken.ExpiresIn = int64(time.Second * time.Since(newToken.Expiry))
		dbToken.RefreshToken = newToken.RefreshToken

		if err = c.db.UpsertAuthorizedUser(dbToken); err != nil {
			return nil, err
		}

		c.SetToken(dbToken)
	}

	return c.config.Client(context.Background(), c.token), nil
}

// UserID returns the ID of the users that authorized this client
// Returns an error if the fitbitClient is not authorized
func (c *FitbitClient) UserID() (*string, error) {
	if c.token == nil || c.userID == nil {
		return nil, errors.New("UserID called, but no user Authorized this client")
	}
	return c.userID, nil
}

// SetToken sets the token inside the FitbitClient. From the types.AuthorizedUser
// to the oauth2.Token representation (privately used).
func (c *FitbitClient) SetToken(token *types.AuthorizedUser) {
	c.token = &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       time.Now().Add(time.Second * time.Duration(token.ExpiresIn)),
		TokenType:    token.TokenType,
	}
	c.userID = &token.UserID
}
