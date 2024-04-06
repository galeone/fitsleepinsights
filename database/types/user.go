package types

import fitbit_pgdb "github.com/galeone/fitbit-pgdb/v3"

type User struct {
	// fitbit_pgdb.AuthorizedUser is already igor-decorated
	fitbit_pgdb.AuthorizedUser
	Dumping bool `sql:"default:true"`
}
