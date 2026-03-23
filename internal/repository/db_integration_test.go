package repository

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func envOrDefault(key, def string) string {
	if v := getEnv(key, ""); v != "" {
		return v
	}
	return def
}

func setMainDBEnvFromTestDB(t *testing.T) {
	t.Helper()

	t.Setenv("DB_HOST", envOrDefault("TEST_DB_HOST", "localhost"))
	t.Setenv("DB_PORT", envOrDefault("TEST_DB_PORT", "5432"))
	t.Setenv("DB_USER", envOrDefault("TEST_DB_USER", "postgres"))
	t.Setenv("DB_PASSWORD", envOrDefault("TEST_DB_PASSWORD", "postgres"))
	t.Setenv("DB_NAME", envOrDefault("TEST_DB_NAME", "restaurant_test"))
}

func openRawTestDB(t *testing.T) *sql.DB {
	t.Helper()

	host := envOrDefault("TEST_DB_HOST", "localhost")
	port := envOrDefault("TEST_DB_PORT", "5432")
	user := envOrDefault("TEST_DB_USER", "postgres")
	password := envOrDefault("TEST_DB_PASSWORD", "postgres")
	dbname := envOrDefault("TEST_DB_NAME", "restaurant_test")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	return db
}

func TestNewDB_Success(t *testing.T) {
	setMainDBEnvFromTestDB(t)

	db, err := NewDB()
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	assert.Equal(t, 25, db.Stats().MaxOpenConnections)
	assert.NoError(t, db.Ping())
}

func TestNewDB_InvalidDatabaseName(t *testing.T) {
	setMainDBEnvFromTestDB(t)
	t.Setenv("DB_NAME", fmt.Sprintf("db_that_does_not_exist_%d", time.Now().UnixNano()))

	db, err := NewDB()
	assert.Error(t, err)
	if db != nil {
		_ = db.Close()
	}
}

func TestRunMigrations_Success(t *testing.T) {
	db := openRawTestDB(t)
	defer db.Close()

	err := RunMigrations(db)
	require.NoError(t, err)

	tables := []string{
		"users",
		"restaurants",
		"menus",
		"menu_items",
		"reservations",
		"orders",
		"order_items",
	}

	for _, tbl := range tables {
		var reg sql.NullString
		err := db.QueryRow(`SELECT to_regclass('public.' || $1)`, tbl).Scan(&reg)
		require.NoError(t, err)
		assert.True(t, reg.Valid, "expected table %s to exist", tbl)
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	db := openRawTestDB(t)
	defer db.Close()

	require.NoError(t, RunMigrations(db))
	require.NoError(t, RunMigrations(db))
}
