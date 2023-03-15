package gormio_zap_test

import (
	"testing"

	"gormio_zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func createTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	// Create a new sqlmock instance.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Create a new GORM instance with the sqlmock driver.
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestLogger(t *testing.T) {
	db, mock := createTestDB(t)

	// Create a new zap observer.
	core, logs := observer.New(zap.DebugLevel)

	// Create a new gormio_zap.Logger instance with the zap observer.
	zapLogger := zap.New(core)
	logger := gormio_zap.New(zapLogger)

	// Set the gormio_zap.Logger as the GORM logger.
	db.Logger = logger

	// Define expected SQL query and mock result.
	mock.ExpectQuery("SELECT \\* FROM users").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "John Doe"))

	type User struct {
		ID   int
		Name string
	}

	// Run a GORM query.
	var users []User
	err := db.Find(&users).Error
	require.NoError(t, err)

	// Check if the logger logs the expected message.
	expectedMessage := "SELECT * FROM users"
	found := false
	for _, log := range logs.All() {
		if log.Message == "trace" && log.ContextMap()["sql"] == expectedMessage {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected log message not found")
}
