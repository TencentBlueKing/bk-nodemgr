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
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

const (
	EnvDockerNetworkKey = "NODEMGR_TEST_DOCKER_NETWORK"

	DockerNetworkBridge = "bridge"
	DockerNetworkHost   = "host"
)

type dockerHostContainerConfig struct {
	Image      string
	Env        map[string]string
	Entrypoint []string
	Cmd        []string
}

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

	startDockerHostContainerWithConfig(t, dockerHostContainerConfig{Image: imageName, Env: env})
}

func startDockerHostContainerWithConfig(t testing.TB, cfg dockerHostContainerConfig) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("create Docker client: %v", err)
	}

	if err := ensureDockerImage(ctx, dockerClient, cfg.Image); err != nil {
		closeDockerClient(t, dockerClient)
		t.Fatalf("ensure Docker image %s: %v", cfg.Image, err)
	}

	created, err := dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image:      cfg.Image,
			Env:        dockerContainerEnv(cfg.Env),
			Entrypoint: strslice.StrSlice(cfg.Entrypoint),
			Cmd:        strslice.StrSlice(cfg.Cmd),
		},
		&container.HostConfig{NetworkMode: container.NetworkMode(DockerNetworkHost)},
		nil,
		nil,
		"",
	)
	if err != nil {
		closeDockerClient(t, dockerClient)
		t.Fatalf("create Docker host-network container %s: %v", cfg.Image, err)
	}

	if err := dockerClient.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		cleanupDockerHostContainer(t, dockerClient, created.ID)
		t.Fatalf("start Docker host-network container %s: %v", cfg.Image, err)
	}

	t.Cleanup(func() {
		cleanupDockerHostContainer(t, dockerClient, created.ID)
	})
}

func cleanupDockerHostContainer(t testing.TB, dockerClient *client.Client, containerID string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil && !errdefs.IsNotFound(err) {
		t.Errorf("remove Docker host-network container %s: %v", containerID, err)
	}
	closeDockerClient(t, dockerClient)
}

func closeDockerClient(t testing.TB, dockerClient *client.Client) {
	t.Helper()

	if err := dockerClient.Close(); err != nil {
		t.Errorf("close Docker client: %v", err)
	}
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
