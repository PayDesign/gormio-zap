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

	mock.ExpectQuery("SELECT \\* FROM `transfer_requests`").WillReturnRows(sqlmock.NewRows([]string{"transfer_request_id", "transfer_status", "amount", "transfer_method", "dry_run", "created_at", "updated_at"}).AddRow(1, 2, 1000, 1, 0, time.Now(), time.Now()))

	type TransferRequest struct {
		TransferRequestID uint64 `gorm:"primaryKey"`
		TransferStatus    uint8
		Amount            uint64
		TransferMethod    uint8
		DryRun            bool
		CreatedAt         time.Time
		UpdatedAt         time.Time
	}
	var expectedTransferRequests = []TransferRequest{{TransferRequestID: 1, TransferStatus: 2, Amount: 1000, TransferMethod: 1, DryRun: false}}

	var transferRequests []TransferRequest
	err := db.Find(&transferRequests).Error
	require.NoError(t, err)

	assert.IsType(t, expectedTransferRequests, transferRequests)

	expectedMessage := "SELECT * FROM `transfer_requests`"

	found := false
	for _, log := range logs.All() {
		if log.Message == "trace" && log.ContextMap()["sql"] == expectedMessage {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected log message not found")
}
