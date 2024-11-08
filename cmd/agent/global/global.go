package global

import (
	"context"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/resourcemanage"
)

var ERManager *resourcemanage.ErrorReleaseManagement

var RootCtx = context.Background()