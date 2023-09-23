// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"fmt"
	"os"

	pgdb "github.com/galeone/fitbit-pgdb/v2"

	_ "github.com/joho/godotenv/autoload"
)

var (
	_connectionString = fmt.Sprintf(
		"host=%s user=%s password=%s port=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	_db           = pgdb.NewPGDB(_connectionString)
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

	// Or, we can use a service account: create a service account and download the key file
	// then set the VAI_SERVICE_ACCOUNT_KEY environment variable to the path of the key file.
	// The service account must have the following roles:
	// - Vertex AI Administrator
	// - Storage Admin

	_vaiLocation          = os.Getenv("VAI_LOCATION")
	_vaiProjectID         = os.Getenv("VAI_PROJECT_ID")
	_vaiServiceAccountKey = os.Getenv("VAI_SERVICE_ACCOUNT_KEY")
	_vaiEndpoint          = fmt.Sprintf("%s-aiplatform.googleapis.com:443", _vaiLocation)
)
