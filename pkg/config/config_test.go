package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	success           = "\u2713"
	failed            = "\u2717"
	validConfig       = "testConfig/test"
	invalidConfigPath = "invalid-config-path/invalid"
)

func TestInitConfig(t *testing.T) {
	t.Run("Should parse config file", func(t *testing.T) {
		t.Logf("\t%s\tStarting InitConfig test with configPath: '%s'", success, validConfig)

		expectedCfg := &Config{
			HTTP: HTTPConfig{
				Port:            "5001",
				ReadTimeout:     10 * time.Second,
				WriteTimeout:    11 * time.Second,
				ShutdownTimeout: 12 * time.Second,
			},
			Mongo: MongoConfig{
				Name:     "some-name",
				URI:      "mongodb://localhost:27017",
				User:     "root",
				Password: "password",
			},
		}

		result, err := Init(validConfig)

		assert.NoError(t, err, fmt.Sprintf("\t%s\t should be able to init configs. Error %v", failed, err))
		t.Logf("\t%s\tNo error returned", success)

		assert.Equal(t, expectedCfg, result, "Initialized config should equal to expected config")
		t.Logf("\t%s\tConfig initialized correctly", success)
	})

	t.Run("Should return err when config file not found", func(t *testing.T) {
		t.Logf("\t%s\tStarting InitConfig test with configPath: '%s'", success, invalidConfigPath)

		result, err := Init(invalidConfigPath)

		assert.Error(t, err, fmt.Sprintf("\t%s\t should be able to init configs. Error %v", failed, err))
		t.Logf("\t%s\tError returned", success)

		assert.Nil(t, result, "Config should be nil")
		t.Logf("\t%s\tConfig is nil", success)
	})
}
