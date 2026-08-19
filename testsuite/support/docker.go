//go:build integration

/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package support

import (
	"context"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
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

	created, err := dockerClient.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:      cfg.Image,
			Env:        dockerContainerEnv(cfg.Env),
			Entrypoint: cfg.Entrypoint,
			Cmd:        cfg.Cmd,
		},
		HostConfig: &container.HostConfig{NetworkMode: container.NetworkMode(DockerNetworkHost)},
	})
	if err != nil {
		closeDockerClient(t, dockerClient)
		t.Fatalf("create Docker host-network container %s: %v", cfg.Image, err)
	}

	if _, err := dockerClient.ContainerStart(ctx, created.ID, client.ContainerStartOptions{}); err != nil {
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

	if _, err := dockerClient.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{Force: true}); err != nil && !errdefs.IsNotFound(err) {
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

	reader, err := dockerClient.ImagePull(ctx, imageName, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	return err
}
