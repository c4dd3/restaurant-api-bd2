package main

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
)

func makeClosableDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	assert.NoError(t, err)
	return db
}

func TestMain_Success(t *testing.T) {
	origRunApp := runApp
	origExit := exitFunc

	called := false
	exited := false

	runApp = func() error {
		called = true
		return nil
	}
	exitFunc = func(code int) {
		exited = true
	}

	defer func() {
		runApp = origRunApp
		exitFunc = origExit
	}()

	main()

	assert.True(t, called)
	assert.False(t, exited)
}

func TestMain_Error(t *testing.T) {
	origRunApp := runApp
	origExit := exitFunc

	exited := false
	exitCode := 0

	runApp = func() error {
		return errors.New("boom")
	}
	exitFunc = func(code int) {
		exited = true
		exitCode = code
	}

	defer func() {
		runApp = origRunApp
		exitFunc = origExit
	}()

	main()

	assert.True(t, exited)
	assert.Equal(t, 1, exitCode)
}

func TestRun_NewDBError(t *testing.T) {
	origLoadEnv := loadEnv
	origNewDB := newDB

	loadEnv = func(filenames ...string) error { return nil }
	newDB = func() (*sql.DB, error) {
		return nil, errors.New("db error")
	}

	defer func() {
		loadEnv = origLoadEnv
		newDB = origNewDB
	}()

	err := run()
	assert.Error(t, err)
}

func TestRun_MigrationError(t *testing.T) {
	origLoadEnv := loadEnv
	origNewDB := newDB
	origRunMigrations := runMigrations

	loadEnv = func(filenames ...string) error { return nil }
	newDB = func() (*sql.DB, error) {
		return makeClosableDB(t), nil
	}
	runMigrations = func(db *sql.DB) error {
		return errors.New("migration error")
	}

	defer func() {
		loadEnv = origLoadEnv
		newDB = origNewDB
		runMigrations = origRunMigrations
	}()

	err := run()
	assert.Error(t, err)
}

func TestRun_Success(t *testing.T) {
	origLoadEnv := loadEnv
	origNewDB := newDB
	origRunMigrations := runMigrations
	origNewJWT := newJWTService
	origSetupRouter := setupRouter
	origRunServer := runServer

	loadCalled := false
	dbCalled := false
	migrateCalled := false
	routerCalled := false
	serverCalled := false

	loadEnv = func(filenames ...string) error {
		loadCalled = true
		return nil
	}
	newDB = func() (*sql.DB, error) {
		dbCalled = true
		return makeClosableDB(t), nil
	}
	runMigrations = func(db *sql.DB) error {
		migrateCalled = true
		return nil
	}
	newJWTService = func() *auth.JWTService {
		return auth.NewJWTService()
	}
	setupRouter = func(
		userRepo handlers.UserRepository,
		restaurantRepo handlers.RestaurantRepository,
		menuRepo handlers.MenuRepository,
		reservationRepo handlers.ReservationRepository,
		orderRepo handlers.OrderRepository,
		jwtSvc *auth.JWTService,
	) *gin.Engine {
		routerCalled = true
		return gin.New()
	}
	runServer = func(run func(...string) error, addr string) error {
		serverCalled = true
		assert.Equal(t, ":8080", addr)
		return nil
	}

	defer func() {
		loadEnv = origLoadEnv
		newDB = origNewDB
		runMigrations = origRunMigrations
		newJWTService = origNewJWT
		setupRouter = origSetupRouter
		runServer = origRunServer
	}()

	err := run()
	assert.NoError(t, err)
	assert.True(t, loadCalled)
	assert.True(t, dbCalled)
	assert.True(t, migrateCalled)
	assert.True(t, routerCalled)
	assert.True(t, serverCalled)
}

func TestRun_ServerError(t *testing.T) {
	origLoadEnv := loadEnv
	origNewDB := newDB
	origRunMigrations := runMigrations
	origNewJWT := newJWTService
	origSetupRouter := setupRouter
	origRunServer := runServer

	loadEnv = func(filenames ...string) error { return nil }
	newDB = func() (*sql.DB, error) {
		return makeClosableDB(t), nil
	}
	runMigrations = func(db *sql.DB) error { return nil }
	newJWTService = func() *auth.JWTService {
		return auth.NewJWTService()
	}
	setupRouter = func(
		userRepo handlers.UserRepository,
		restaurantRepo handlers.RestaurantRepository,
		menuRepo handlers.MenuRepository,
		reservationRepo handlers.ReservationRepository,
		orderRepo handlers.OrderRepository,
		jwtSvc *auth.JWTService,
	) *gin.Engine {
		return gin.New()
	}
	runServer = func(run func(...string) error, addr string) error {
		return errors.New("server error")
	}

	defer func() {
		loadEnv = origLoadEnv
		newDB = origNewDB
		runMigrations = origRunMigrations
		newJWTService = origNewJWT
		setupRouter = origSetupRouter
		runServer = origRunServer
	}()

	err := run()
	assert.Error(t, err)
}
