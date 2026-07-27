//go:build integration

package support

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

func uniqueName(prefix string) string {
	id := strings.ReplaceAll(uuid.NewString(), "-", "")
	return prefix + "_" + time.Now().UTC().Format("20060102150405") + "_" + id[:12]
}
