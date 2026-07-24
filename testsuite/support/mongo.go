//go:build integration

package support

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultMongoImage = "docker.io/bitnamilegacy/mongodb:6.0.10-debian-11-r8"

// RequireMongoClient returns a MongoDB client for integration tests.
func RequireMongoClient(t testing.TB) *mongo.Client {
	t.Helper()

	client, _ := RequireMongoDatabase(t)
	return client
}

// RequireMongoDatabase returns a MongoDB client and isolated database for integration tests.
func RequireMongoDatabase(t testing.TB) (*mongo.Client, *mongo.Database) {
	t.Helper()

	src, values := resolveSource(t)
	switch src {
	case sourceEnv:
		return requireMongoFromEnv(t, mongoEnvFromValues(t, values))
	case sourceDocker:
		return requireMongoFromDocker(t, values)
	default:
		t.Fatalf("unsupported Mongo source %q", src)
		return nil, nil
	}
}

func requireMongoFromEnv(t testing.TB, cfg mongoEnvConfig) (*mongo.Client, *mongo.Database) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, &options.ClientOptions{
		Hosts: []string{cfg.Address},
		Auth: &options.Credential{
			Username:      cfg.User,
			Password:      cfg.Password,
			AuthSource:    cfg.AuthSource,
			AuthMechanism: cfg.AuthMechanism,
		},
	})
	if err != nil {
		t.Fatalf("connect MongoDB from %s: %v", integrationEnvPath, err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		t.Fatalf("ping MongoDB from %s: %v", integrationEnvPath, err)
	}

	databaseName := uniqueName(cfg.Database)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = client.Database(databaseName).Drop(ctx)
		_ = client.Disconnect(ctx)
	})

	return client, client.Database(databaseName)
}

func requireMongoFromDocker(t testing.TB, values map[string]string) (*mongo.Client, *mongo.Database) {
	t.Helper()

	imageName := mongoImage(values)
	switch dockerNetwork(t, values) {
	case DockerNetworkBridge:
		return requireMongoFromDockerBridge(t, imageName)
	case DockerNetworkHost:
		return requireMongoFromDockerHost(t, imageName)
	default:
		t.Fatalf("unsupported Docker network")
		return nil, nil
	}
}

func mongoImage(values map[string]string) string {
	if imageName := configValue(values, EnvMongoImageKey); imageName != "" {
		return imageName
	}
	return defaultMongoImage
}

func requireMongoFromDockerBridge(t testing.TB, imageName string) (*mongo.Client, *mongo.Database) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("start MongoDB container: %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := mongodb.Run(ctx, imageName)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start MongoDB container: %v", err)
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get MongoDB container connection string: %v", err)
	}

	return requireMongoFromURI(t, uri)
}

func requireMongoFromDockerHost(t testing.TB, imageName string) (*mongo.Client, *mongo.Database) {
	t.Helper()

	startDockerHostContainer(t, imageName, nil)
	return requireMongoFromURI(t, "mongodb://127.0.0.1:27017")
}

func requireMongoFromURI(t testing.TB, uri string) (*mongo.Client, *mongo.Database) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect MongoDB container: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		t.Fatalf("ping MongoDB container: %v", err)
	}

	databaseName := uniqueName("nodemgr_test")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = client.Database(databaseName).Drop(ctx)
		_ = client.Disconnect(ctx)
	})

	return client, client.Database(databaseName)
}
