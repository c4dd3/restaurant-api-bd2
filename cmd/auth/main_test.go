package main

import (
	"database/sql"
	"errors"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"restaurant-api/internal/auth"
)

func makeClosableDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	assert.NoError(t, err)
	return db
}

func TestAuthMain_Success(t *testing.T) {
	origRunApp, origExit := runApp, exitFunc
	called, exited := false, false
	runApp = func() error { called = true; return nil }
	exitFunc = func(code int) { exited = true }
	defer func() { runApp = origRunApp; exitFunc = origExit }()
	main()
	assert.True(t, called)
	assert.False(t, exited)
}

func TestAuthMain_Error(t *testing.T) {
	origRunApp, origExit := runApp, exitFunc
	exitCode := 0
	runApp = func() error { return errors.New("boom") }
	exitFunc = func(code int) { exitCode = code }
	defer func() { runApp = origRunApp; exitFunc = origExit }()
	main()
	assert.Equal(t, 1, exitCode)
}

func TestAuthRun_DBError(t *testing.T) {
	origLoadEnv, origNewDB := loadEnv, newDB
	loadEnv = func(filenames ...string) error { return nil }
	newDB = func() (*sql.DB, error) { return nil, errors.New("db error") }
	defer func() { loadEnv = origLoadEnv; newDB = origNewDB }()
	assert.Error(t, run())
}

func TestAuthRun_MigrationError(t *testing.T) {
	origLoadEnv, origNewDB, origRunMigrations := loadEnv, newDB, runMigrations
	loadEnv = func(filenames ...string) error { return nil }
	newDB = func() (*sql.DB, error) { return makeClosableDB(t), nil }
	runMigrations = func(db *sql.DB) error { return errors.New("migration error") }
	defer func() {
		loadEnv = origLoadEnv
		newDB = origNewDB
		runMigrations = origRunMigrations
	}()
	assert.Error(t, run())
}

func TestAuthRun_Success(t *testing.T) {
	origLoadEnv := loadEnv
	origNewDB := newDB
	origRunMigrations := runMigrations
	origNewJWT := newJWTService
	origRunServer := runServer

	serverCalled := false
	loadEnv = func(filenames ...string) error { return nil }
	newDB = func() (*sql.DB, error) { return makeClosableDB(t), nil }
	runMigrations = func(db *sql.DB) error { return nil }
	newJWTService = func() *auth.JWTService { return auth.NewJWTService() }
	runServer = func(run func(...string) error, addr string) error {
		serverCalled = true
		assert.Equal(t, ":8081", addr)
		return nil
	}
	defer func() {
		loadEnv = origLoadEnv
		newDB = origNewDB
		runMigrations = origRunMigrations
		newJWTService = origNewJWT
		runServer = origRunServer
	}()

	assert.NoError(t, run())
	assert.True(t, serverCalled)
}

func TestAuthRun_CustomPort(t *testing.T) {
	origLoadEnv := loadEnv
	origNewDB := newDB
	origRunMigrations := runMigrations
	origNewJWT := newJWTService
	origRunServer := runServer

	t.Setenv("AUTH_PORT", "9090")
	loadEnv = func(filenames ...string) error { return nil }
	newDB = func() (*sql.DB, error) { return makeClosableDB(t), nil }
	runMigrations = func(db *sql.DB) error { return nil }
	newJWTService = func() *auth.JWTService { return auth.NewJWTService() }
	runServer = func(run func(...string) error, addr string) error {
		assert.Equal(t, ":9090", addr)
		return nil
	}
	defer func() {
		loadEnv = origLoadEnv
		newDB = origNewDB
		runMigrations = origRunMigrations
		newJWTService = origNewJWT
		runServer = origRunServer
	}()

	assert.NoError(t, run())
}

func TestAuthRun_ServerError(t *testing.T) {
	origLoadEnv := loadEnv
	origNewDB := newDB
	origRunMigrations := runMigrations
	origNewJWT := newJWTService
	origRunServer := runServer

	loadEnv = func(filenames ...string) error { return nil }
	newDB = func() (*sql.DB, error) { return makeClosableDB(t), nil }
	runMigrations = func(db *sql.DB) error { return nil }
	newJWTService = func() *auth.JWTService { return auth.NewJWTService() }
	runServer = func(run func(...string) error, addr string) error { return errors.New("server error") }
	defer func() {
		loadEnv = origLoadEnv
		newDB = origNewDB
		runMigrations = origRunMigrations
		newJWTService = origNewJWT
		runServer = origRunServer
	}()

	assert.Error(t, run())
}
