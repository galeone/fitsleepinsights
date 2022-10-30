package fitbit

import (
	"github.com/galeone/sleepbit/fitbit/types"
)

// Storage is the interface to implement for implementing the persistence
// layer required by the fitbit client.
type Storage interface {
	// InsertAuhorizingUser creates a new AuthorizingUser.
	// An AuthorizingUser is an user in the process of giving the
	// authorization to the fitbit-API-based application.
	InsertAuhorizingUser(*types.AuthorizingUser) error

	// UpsertAuthorizedUser creates or updates an AuthorizedUser.
	// It creates the user if it's not present in the storage. It updates its
	// attributes if the very same user is already present in the storage.
	// An AuthorizedUser is an user that completed the Authorization phase
	// that means it transitioned from the state of AuthorizingUser to the
	// state of AuthorizedUser.
	UpsertAuthorizedUser(*types.AuthorizedUser) error

	// AuthorizedUser returns a pointer to types.AuthorizedUser
	// given the accessToken.
	AuthorizedUser(accessToken string) (*types.AuthorizedUser, error)

	// AuthorizingUser returns a pointer to the types.AuthorizingUser
	// given it's unique identifier (any ID - in DB implementation often
	// a primary key).
	AuthorizingUser(id string) (*types.AuthorizingUser, error)
}
