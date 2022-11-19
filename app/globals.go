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
)
