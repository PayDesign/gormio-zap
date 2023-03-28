package gormio_zap_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	gormio_zap "github.com/PayDesign/gormio-zap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func createTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestLogger(t *testing.T) {
	db, mock := createTestDB(t)

	core, logs := observer.New(zap.DebugLevel)

	zapLogger := zap.New(core)

	logger := gormio_zap.New(zapLogger)

	db.Logger = logger

	mock.ExpectQuery("SELECT \\* FROM `users`").WillReturnRows(sqlmock.NewRows([]string{"user_id", "name", "created_at", "updated_at"}).AddRow(1, "test", time.Now(), time.Now()))

	type User struct {
		UserID    uint64 `gorm:"primaryKey"`
		Name      string
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	var expectedUsers = []User{{UserID: 1, Name: "test"}}

	var users []User
	err := db.Find(&users).Error
	require.NoError(t, err)

	assert.IsType(t, expectedUsers, users)

	expectedMessage := "SELECT * FROM `users`"

	found := false
	for _, log := range logs.All() {
		if log.Message == "trace" && log.ContextMap()["sql"] == expectedMessage {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected log message not found")
}
