//go:build integration

package support

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

const (
	EnvSourceKey = "NODEMGR_TEST_ENV_SOURCE"

	SourceAuto   = "auto"
	SourceEnv    = "env"
	SourceDocker = "docker"

	EnvMongoImageKey         = "NODEMGR_TEST_MONGO_IMAGE"
	EnvMongoAddressKey       = "NODEMGR_TEST_MONGO_ADDRESS"
	EnvMongoUserKey          = "NODEMGR_TEST_MONGO_USER"
	EnvMongoPasswordKey      = "NODEMGR_TEST_MONGO_PASSWORD"
	EnvMongoAuthSourceKey    = "NODEMGR_TEST_MONGO_AUTH_SOURCE"
	EnvMongoAuthMechanismKey = "NODEMGR_TEST_MONGO_AUTH_MECHANISM"
	EnvMongoDatabaseKey      = "NODEMGR_TEST_MONGO_DATABASE"

	EnvRedisImageKey    = "NODEMGR_TEST_REDIS_IMAGE"
	EnvRedisAddressKey  = "NODEMGR_TEST_REDIS_ADDRESS"
	EnvRedisUsernameKey = "NODEMGR_TEST_REDIS_USERNAME"
	EnvRedisPasswordKey = "NODEMGR_TEST_REDIS_PASSWORD"
	EnvRedisDBKey       = "NODEMGR_TEST_REDIS_DB"

	integrationEnvPath = "testsuite/integration.env"
)

type mongoEnvConfig struct {
	Address       string
	User          string
	Password      string
	AuthSource    string
	AuthMechanism string
	Database      string
}

type redisEnvConfig struct {
	Address  string
	Username string
	Password string
	DB       int
}

func readEnvFile(t testing.TB) (map[string]string, bool) {
	t.Helper()

	path := integrationEnvFile(t)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		t.Fatalf("stat integration env file %s: %v", path, err)
	}

	values, err := godotenv.Read(path)
	if err != nil {
		t.Fatalf("read integration env file %s: %v", path, err)
	}
	return values, true
}

func mongoEnvFromValues(t testing.TB, values map[string]string) mongoEnvConfig {
	t.Helper()

	return mongoEnvConfig{
		Address:       requiredConfigValue(t, values, EnvMongoAddressKey),
		User:          requiredConfigValue(t, values, EnvMongoUserKey),
		Password:      requiredConfigValue(t, values, EnvMongoPasswordKey),
		AuthSource:    requiredConfigValue(t, values, EnvMongoAuthSourceKey),
		AuthMechanism: requiredConfigValue(t, values, EnvMongoAuthMechanismKey),
		Database:      requiredConfigValue(t, values, EnvMongoDatabaseKey),
	}
}

func redisEnvFromValues(t testing.TB, values map[string]string) redisEnvConfig {
	t.Helper()

	redisDB, err := strconv.Atoi(requiredConfigValue(t, values, EnvRedisDBKey))
	if err != nil {
		t.Fatalf("parse %s from %s: %v", EnvRedisDBKey, integrationEnvPath, err)
	}

	return redisEnvConfig{
		Address:  requiredConfigValue(t, values, EnvRedisAddressKey),
		Username: configValue(values, EnvRedisUsernameKey),
		Password: configValue(values, EnvRedisPasswordKey),
		DB:       redisDB,
	}
}

func integrationEnvFile(t testing.TB) string {
	t.Helper()

	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("find repo root: %v", err)
	}
	return filepath.Join(root, integrationEnvPath)
}

func requiredConfigValue(t testing.TB, values map[string]string, key string) string {
	t.Helper()

	value := configValue(values, key)
	if value == "" {
		t.Fatalf("%s is required in %s", key, integrationEnvPath)
	}
	return value
}

func configValue(values map[string]string, key string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return strings.TrimSpace(values[key])
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}
