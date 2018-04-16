package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func SetupDatabaseForTests(t *testing.T) {
	SetupTestDatabase()
	tables := ShowTables()
	assert.Equal(t, 3, len(tables))
}

func TestDatabase(t *testing.T) {
	SetupDatabaseForTests(t)
	UsersTest(t)
	MessagesTest(t)
	SessionsTest(t)
}
