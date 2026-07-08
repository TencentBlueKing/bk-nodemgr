package node

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// DiagnoseAgentVersion runs the agent version diagnostic and logs the raw result.
func DiagnoseAgentVersion(ctx context.Context, step logger.Step, scene string, handler agenthandler.IAgentHandler) error {
	diagnostic, err := handler.Process().DiagnoseVersion(ctx)
	if err != nil {
		if diagnostic != nil {
			logger.Errorf(step, "failed to diagnose agent version: scene=%s work_dir(%s) executable(%s) args(%v) stdout_raw(%s) stderr_raw(%s): %v", scene, diagnostic.WorkDir, diagnostic.Executable, diagnostic.Args, diagnostic.Stdout, diagnostic.Stderr, err)
		} else {
			logger.Errorf(step, "failed to diagnose agent version: scene=%s: %v", scene, err)
		}
		return fmt.Errorf("failed to diagnose agent version for %s: %w", scene, err)
	}

	logger.Infof(step, "agent version diagnostic: scene=%s work_dir(%s) executable(%s) args(%v) stdout_raw(%s) stderr_raw(%s)", scene, diagnostic.WorkDir, diagnostic.Executable, diagnostic.Args, diagnostic.Stdout, diagnostic.Stderr)
	return nil
}
