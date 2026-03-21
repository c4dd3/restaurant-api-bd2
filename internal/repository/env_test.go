package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv_ReturnsFallbackWhenEmpty(t *testing.T) {
	t.Setenv("TEST_ENV_EMPTY", "")
	got := getEnv("TEST_ENV_EMPTY", "fallback-value")
	assert.Equal(t, "fallback-value", got)
}

func TestGetEnv_ReturnsValueWhenSet(t *testing.T) {
	t.Setenv("TEST_ENV_VALUE", "real-value")
	got := getEnv("TEST_ENV_VALUE", "fallback-value")
	assert.Equal(t, "real-value", got)
}
