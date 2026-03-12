package etcddiscover

import (
	"crypto/tls"
	"errors"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewEtcdClientConfigInjectsUnifiedLogger(t *testing.T) {
	tlsConf := &tls.Config{MinVersion: tls.VersionTLS12}
	etcdConf := &config.Etcd{
		Endpoints: []string{"https://127.0.0.1:2379"},
		Username:  "user",
		Password:  "pass",
	}

	cfg := newEtcdClientConfig(etcdConf, tlsConf)

	if cfg.Logger == nil {
		t.Fatal("expected etcd client config logger to be injected")
	}
	if cfg.LogConfig != nil {
		t.Fatal("expected etcd client config to reuse injected logger instead of LogConfig")
	}
	if got, want := cfg.Endpoints, etcdConf.Endpoints; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("unexpected endpoints: got %v, want %v", got, want)
	}
	if cfg.Username != etcdConf.Username {
		t.Fatalf("unexpected username: got %q, want %q", cfg.Username, etcdConf.Username)
	}
	if cfg.Password != etcdConf.Password {
		t.Fatalf("unexpected password: got %q, want %q", cfg.Password, etcdConf.Password)
	}
	if cfg.TLS != tlsConf {
		t.Fatal("expected TLS config to be forwarded unchanged")
	}
}

func TestMapEtcdLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		input zapcore.Level
		want  logger.Level
	}{
		{name: "debug", input: zapcore.DebugLevel, want: logger.LevelDebug},
		{name: "info", input: zapcore.InfoLevel, want: logger.LevelInfo},
		{name: "warn", input: zapcore.WarnLevel, want: logger.LevelWarn},
		{name: "error", input: zapcore.ErrorLevel, want: logger.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapEtcdLogLevel(tt.input); got != tt.want {
				t.Fatalf("mapEtcdLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEtcdLoggerCorePreservesDiagnosticFields(t *testing.T) {
	var (
		gotLevel logger.Level
		gotMsg   string
		gotKVs   []interface{}
	)

	core := &etcdLoggerCore{
		log: func(level logger.Level, msg string, kvs ...interface{}) {
			gotLevel = level
			gotMsg = msg
			gotKVs = append([]interface{}(nil), kvs...)
		},
		fields: []zapcore.Field{
			zap.String("component", "retry_interceptor"),
		},
	}

	err := core.Write(zapcore.Entry{
		Level:   zapcore.WarnLevel,
		Message: "retrying unary invoker failed",
		Caller: zapcore.EntryCaller{
			Defined: true,
			File:    "go.etcd.io/etcd/client/v3/retry_interceptor.go",
			Line:    321,
		},
	}, []zapcore.Field{
		zap.Error(errors.New("boom")),
		zap.String("target", "etcd://cluster/member"),
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if gotLevel != logger.LevelWarn {
		t.Fatalf("unexpected mapped level: got %v, want %v", gotLevel, logger.LevelWarn)
	}
	if gotMsg != "retrying unary invoker failed" {
		t.Fatalf("unexpected message: got %q", gotMsg)
	}

	kvMap := make(map[string]interface{}, len(gotKVs)/2)
	for i := 0; i+1 < len(gotKVs); i += 2 {
		key, ok := gotKVs[i].(string)
		if !ok {
			t.Fatalf("expected string key at index %d, got %T", i, gotKVs[i])
		}
		kvMap[key] = gotKVs[i+1]
	}

	if kvMap["level"] != zapcore.WarnLevel.String() {
		t.Fatalf("unexpected level field: got %v", kvMap["level"])
	}
	if kvMap["caller"] != "v3/retry_interceptor.go:321" {
		t.Fatalf("unexpected caller field: got %v", kvMap["caller"])
	}
	if kvMap["component"] != "retry_interceptor" {
		t.Fatalf("unexpected component field: got %v", kvMap["component"])
	}
	if kvMap["target"] != "etcd://cluster/member" {
		t.Fatalf("unexpected target field: got %v", kvMap["target"])
	}
	if kvMap["error"] != "boom" {
		t.Fatalf("unexpected error field: got %v", kvMap["error"])
	}
}
