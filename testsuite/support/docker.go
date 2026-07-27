//go:build integration

package support

import (
	"context"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

const (
	EnvDockerNetworkKey = "NODEMGR_TEST_DOCKER_NETWORK"

	DockerNetworkBridge = "bridge"
	DockerNetworkHost   = "host"
)

func dockerNetwork(t testing.TB, values map[string]string) string {
	t.Helper()

	network := configValue(values, EnvDockerNetworkKey)
	if network == "" {
		return DockerNetworkBridge
	}

	switch network {
	case DockerNetworkBridge, DockerNetworkHost:
		return network
	default:
		t.Fatalf("unsupported %s %q, want bridge or host", EnvDockerNetworkKey, network)
		return ""
	}
}

func startDockerHostContainer(t testing.TB, imageName string, env map[string]string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("create Docker client: %v", err)
	}

	if err := ensureDockerImage(ctx, dockerClient, imageName); err != nil {
		_ = dockerClient.Close()
		t.Fatalf("ensure Docker image %s: %v", imageName, err)
	}

	created, err := dockerClient.ContainerCreate(
		ctx,
		&container.Config{Image: imageName, Env: dockerContainerEnv(env)},
		&container.HostConfig{NetworkMode: container.NetworkMode(DockerNetworkHost)},
		nil,
		nil,
		"",
	)
	if err != nil {
		_ = dockerClient.Close()
		t.Fatalf("create Docker host-network container %s: %v", imageName, err)
	}

	if err := dockerClient.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		_ = dockerClient.ContainerRemove(context.Background(), created.ID, container.RemoveOptions{Force: true})
		_ = dockerClient.Close()
		t.Fatalf("start Docker host-network container %s: %v", imageName, err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = dockerClient.ContainerRemove(ctx, created.ID, container.RemoveOptions{Force: true})
		_ = dockerClient.Close()
	})
}

func dockerContainerEnv(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}

	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, key+"="+env[key])
	}
	return values
}

func ensureDockerImage(ctx context.Context, dockerClient *client.Client, imageName string) error {
	_, err := dockerClient.ImageInspect(ctx, imageName)
	if err == nil {
		return nil
	}
	if !errdefs.IsNotFound(err) {
		return err
	}

	reader, err := dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	return err
}
