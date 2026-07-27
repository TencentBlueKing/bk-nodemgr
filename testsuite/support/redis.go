//go:build integration

package support

import (
	"context"
	"testing"
	"time"

	goredislib "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

const defaultRedisImage = "docker.io/bitnamilegacy/redis:7.0.12-debian-11-r34"

// RequireRedisClient returns a Redis client for integration tests.
func RequireRedisClient(t testing.TB) goredislib.UniversalClient {
	t.Helper()

	client, _ := RequireRedisClientWithKeyPrefix(t)
	return client
}

// RequireRedisClientWithKeyPrefix returns a Redis client and isolated key prefix for integration tests.
func RequireRedisClientWithKeyPrefix(t testing.TB) (goredislib.UniversalClient, string) {
	t.Helper()

	src, values := resolveSource(t)
	switch src {
	case sourceEnv:
		return requireRedisFromEnv(t, redisEnvFromValues(t, values))
	case sourceDocker:
		return requireRedisFromDocker(t, values)
	default:
		t.Fatalf("unsupported Redis source %q", src)
		return nil, ""
	}
}

func requireRedisFromEnv(t testing.TB, cfg redisEnvConfig) (goredislib.UniversalClient, string) {
	t.Helper()

	client := goredislib.NewClient(&goredislib.Options{
		Addr:     cfg.Address,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		t.Fatalf("ping Redis from %s: %v", integrationEnvPath, err)
	}

	prefix := uniqueName("nodemgr_test")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = cleanupRedisPrefix(ctx, client, prefix)
		_ = client.Close()
	})

	return client, prefix
}

func requireRedisFromDocker(t testing.TB, values map[string]string) (goredislib.UniversalClient, string) {
	t.Helper()

	imageName := redisImage(values)
	switch dockerNetwork(t, values) {
	case DockerNetworkBridge:
		return requireRedisFromDockerBridge(t, imageName)
	case DockerNetworkHost:
		return requireRedisFromDockerHost(t, imageName)
	default:
		t.Fatalf("unsupported Docker network")
		return nil, ""
	}
}

func redisImage(values map[string]string) string {
	if imageName := configValue(values, EnvRedisImageKey); imageName != "" {
		return imageName
	}
	return defaultRedisImage
}

func requireRedisFromDockerBridge(t testing.TB, imageName string) (goredislib.UniversalClient, string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("start Redis container: %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := tcredis.Run(ctx, imageName, redisDockerEnv())
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start Redis container: %v", err)
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get Redis container connection string: %v", err)
	}

	return requireRedisFromURI(t, uri)
}

func requireRedisFromDockerHost(t testing.TB, imageName string) (goredislib.UniversalClient, string) {
	t.Helper()

	startDockerHostContainer(t, imageName, redisDockerEnvValues())
	return requireRedisFromURI(t, "redis://127.0.0.1:6379")
}

func redisDockerEnv() testcontainers.ContainerCustomizer {
	return testcontainers.WithEnv(redisDockerEnvValues())
}

func redisDockerEnvValues() map[string]string {
	return map[string]string{"ALLOW_EMPTY_PASSWORD": "yes"}
}

func requireRedisFromURI(t testing.TB, uri string) (goredislib.UniversalClient, string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	options, err := goredislib.ParseURL(uri)
	if err != nil {
		t.Fatalf("parse Redis container connection string: %v", err)
	}

	client := goredislib.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		t.Fatalf("ping Redis container: %v", err)
	}

	prefix := uniqueName("nodemgr_test")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = cleanupRedisPrefix(ctx, client, prefix)
		_ = client.Close()
	})

	return client, prefix
}

func cleanupRedisPrefix(ctx context.Context, client goredislib.UniversalClient, prefix string) error {
	var cursor uint64
	pattern := prefix + "*"
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		cursor = nextCursor

		if len(keys) > 0 {
			if err := client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		if cursor == 0 {
			return nil
		}
	}
}
