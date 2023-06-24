package database

import (
	"fmt"
	"os"
)

var (
	_connectionString = fmt.Sprintf(
		"host=%s user=%s password=%s port=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
)

const (
	NewUsersChannel string = "new_users"
)
