//go:build integration

package support

import "testing"

type source string

const (
	sourceAuto   source = SourceAuto
	sourceEnv    source = SourceEnv
	sourceDocker source = SourceDocker
)

func resolveSource(t testing.TB) (source, map[string]string) {
	t.Helper()

	values, hasEnv := readEnvFile(t)
	configured := configValue(values, EnvSourceKey)
	if configured == "" {
		configured = SourceAuto
	}

	switch source(configured) {
	case sourceAuto:
		if hasEnv {
			return sourceEnv, values
		}
		return sourceDocker, nil
	case sourceEnv:
		if !hasEnv {
			t.Fatalf("%s=env requires %s", EnvSourceKey, integrationEnvPath)
		}
		return sourceEnv, values
	case sourceDocker:
		return sourceDocker, values
	default:
		t.Fatalf("unsupported %s %q, want auto, env, or docker", EnvSourceKey, configured)
		return "", nil
	}
}
