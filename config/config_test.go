package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("../config.env")

	assert.NoError(t, err)
	assert.Equal(t, config.HttpPort, "8080")
}
