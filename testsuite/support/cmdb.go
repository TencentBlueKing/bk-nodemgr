//go:build integration

package support

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultCMDBImage           = "ccr.ccs.tencentyun.com/bk.io/cmdb-standalone@sha256:1987ff3dcf56debc6961d16e5669404c02816e7e31e7f7711fb3e054a1f06302"
	defaultCMDBSupplierAccount = "0"
	defaultCMDBVirtualUser     = "admin"
	defaultCMDBTenantID        = "default"
	defaultCMDBBridgeBindIP    = "0.0.0.0"
	cmdbContainerPort          = "8080/tcp"
	cmdbReadyPath              = "/healthz"
	cmdbDockerStartupTimeout   = 5 * time.Minute
	cmdbEnvStartupTimeout      = 10 * time.Second
)

// CMDBTarget describes a CMDB target for integration tests.
type CMDBTarget struct {
	Endpoint        string
	SupplierAccount string
	VirtualUser     string
	AppCode         string
	TenantID        string
}

// RequireCMDBTarget returns a ready CMDB target for integration tests.
func RequireCMDBTarget(t testing.TB) CMDBTarget {
	t.Helper()

	src, values := resolveSource(t)
	switch src {
	case sourceEnv:
		return requireCMDBFromEnv(t, cmdbEnvFromValues(t, values))
	case sourceDocker:
		return requireCMDBFromDocker(t, values)
	default:
		t.Fatalf("unsupported CMDB source %q", src)
		return CMDBTarget{}
	}
}

func requireCMDBFromEnv(t testing.TB, cfg cmdbTargetConfig) CMDBTarget {
	t.Helper()

	endpoint := normalizeCMDBEndpoint(t, cfg.Endpoint)
	waitForCMDBReady(t, endpoint, cmdbEnvStartupTimeout)

	return CMDBTarget{
		Endpoint:        endpoint,
		SupplierAccount: cfg.SupplierAccount,
		VirtualUser:     cfg.VirtualUser,
		AppCode:         cfg.AppCode,
		TenantID:        cfg.TenantID,
	}
}

func requireCMDBFromDocker(t testing.TB, values map[string]string) CMDBTarget {
	t.Helper()

	target := cmdbDockerTargetFromValues(t, values)
	imageName := cmdbImage(values)

	switch dockerNetwork(t, values) {
	case DockerNetworkBridge:
		return requireCMDBFromDockerBridge(t, imageName, target, values)
	case DockerNetworkHost:
		return requireCMDBFromDockerHost(t, imageName, target, values)
	default:
		t.Fatalf("unsupported Docker network")
		return CMDBTarget{}
	}
}

func cmdbDockerTargetFromValues(t testing.TB, values map[string]string) CMDBTarget {
	t.Helper()

	supplierAccount := configValue(values, EnvCMDBSupplierAccountKey)
	if supplierAccount == "" {
		supplierAccount = defaultCMDBSupplierAccount
	}
	virtualUser := configValue(values, EnvCMDBVirtualUserKey)
	if virtualUser == "" {
		virtualUser = defaultCMDBVirtualUser
	}

	return CMDBTarget{
		SupplierAccount: supplierAccount,
		VirtualUser:     virtualUser,
		AppCode:         configValue(values, EnvCMDBAppCodeKey),
		TenantID:        cmdbTenantID(values),
	}
}

func cmdbTenantID(values map[string]string) string {
	if tenantID := configValue(values, EnvCMDBTenantIDKey); tenantID != "" {
		return tenantID
	}
	return defaultCMDBTenantID
}

func cmdbImage(values map[string]string) string {
	if imageName := configValue(values, EnvCMDBImageKey); imageName != "" {
		return imageName
	}
	return defaultCMDBImage
}

func requireCMDBFromDockerBridge(t testing.TB, imageName string, target CMDBTarget, values map[string]string) CMDBTarget {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("start CMDB container: %v", r)
		}
	}()

	hostIP := cmdbHostIP(values)
	if hostIP == "" {
		hostIP = defaultCMDBBridgeBindIP
	}
	validateCMDBHostIP(t, hostIP)

	ctx, cancel := context.WithTimeout(context.Background(), cmdbDockerStartupTimeout)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        imageName,
			Entrypoint:   []string{"/bin/sh"},
			Cmd:          cmdbDockerCommand(hostIP),
			ExposedPorts: []string{cmdbContainerPort},
			WaitingFor: wait.ForHTTP(cmdbReadyPath).
				WithPort(cmdbContainerPort).
				WithStartupTimeout(cmdbDockerStartupTimeout),
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start CMDB container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("get CMDB container host: %v", err)
	}
	mappedPort, err := container.MappedPort(ctx, cmdbContainerPort)
	if err != nil {
		t.Fatalf("get CMDB container mapped port: %v", err)
	}

	target.Endpoint = "http://" + net.JoinHostPort(host, mappedPort.Port())
	waitForCMDBReady(t, target.Endpoint, cmdbEnvStartupTimeout)
	return target
}

func requireCMDBFromDockerHost(t testing.TB, imageName string, target CMDBTarget, values map[string]string) CMDBTarget {
	t.Helper()

	hostIP := cmdbHostIP(values)
	if hostIP == "" {
		hostIP = "127.0.0.1"
	}
	validateCMDBHostIP(t, hostIP)

	startDockerHostContainerWithConfig(t, dockerHostContainerConfig{
		Image:      imageName,
		Entrypoint: []string{"/bin/sh"},
		Cmd:        cmdbDockerCommand(hostIP),
	})

	target.Endpoint = "http://" + net.JoinHostPort(hostIP, "8080")
	waitForCMDBReady(t, target.Endpoint, cmdbDockerStartupTimeout)
	return target
}

func cmdbHostIP(values map[string]string) string {
	return configValue(values, EnvCMDBHostIPKey)
}

func validateCMDBHostIP(t testing.TB, hostIP string) {
	t.Helper()

	if net.ParseIP(hostIP) == nil {
		t.Fatalf("invalid %s %q, want IPv4 or IPv6 address", EnvCMDBHostIPKey, hostIP)
	}
}

func cmdbDockerCommand(hostIP string) []string {
	return []string{
		"-c",
		fmt.Sprintf("sed -i 's/ip=127.0.0.1/ip=%s/' /data/run.sh && /data/run.sh", hostIP),
	}
}

func normalizeCMDBEndpoint(t testing.TB, endpoint string) string {
	t.Helper()

	parsed, err := url.Parse(endpoint)
	if err != nil {
		t.Fatalf("parse %s from %s: %v", EnvCMDBEndpointKey, integrationEnvPath, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		t.Fatalf("unsupported %s scheme %q, want http or https", EnvCMDBEndpointKey, parsed.Scheme)
	}
	if parsed.User != nil {
		t.Fatalf("%s must not include user info", EnvCMDBEndpointKey)
	}
	if parsed.Hostname() == "" {
		t.Fatalf("%s must include host", EnvCMDBEndpointKey)
	}
	if parsed.RawQuery != "" {
		t.Fatalf("%s must be CMDB apiserver origin without query", EnvCMDBEndpointKey)
	}
	if parsed.Fragment != "" {
		t.Fatalf("%s must be CMDB apiserver origin without fragment", EnvCMDBEndpointKey)
	}
	if parsed.Path != "" && parsed.Path != "/" {
		t.Fatalf("%s must be CMDB apiserver origin without path", EnvCMDBEndpointKey)
	}

	parsed.Path = ""
	parsed.RawPath = ""
	return parsed.String()
}

func waitForCMDBReady(t testing.TB, endpoint string, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	readyURL := endpoint + cmdbReadyPath
	client := &http.Client{Timeout: 5 * time.Second}
	var lastErr error

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, readyURL, nil)
		if err != nil {
			t.Fatalf("create CMDB readiness request: %v", err)
		}

		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
			lastErr = fmt.Errorf("status %s", resp.Status)
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			t.Fatalf("wait for CMDB readiness at %s: %v", endpoint, lastErr)
		case <-time.After(500 * time.Millisecond):
		}
	}
}
