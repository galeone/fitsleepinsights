// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"os"

	pgdb "github.com/galeone/fitbit-pgdb"

	_ "github.com/joho/godotenv/autoload"
)

var (
	_db           = pgdb.NewPGDB()
	_clientID     = os.Getenv("FITBIT_CLIENT_ID")
	_clientSecret = os.Getenv("FITBIT_CLIENT_SECRET")
	_redirectURL  = os.Getenv("FITBIT_REDIRECT_URL")
	_domain       = os.Getenv("DOMAIN")

	// VertexAI:
	// prerequisite
	// ```
	// gcloud auth application-default login
	// ```
	// This creates a file (application default credentials) in a well-known location
	// and the sdk uses this location to setup the account & project
	//
	// info: https://stackoverflow.com/a/52247638/2891324
	_vaiLocation   = os.Getenv("VAI_LOCATION")
	_vaiProjectID  = os.Getenv("VAI_PROJECT_ID")
	_vaiEndpointID = os.Getenv("VAI_ENDPOINT_ID")
)
