package integration

import (
	"testing"
)

// TestAppStartup verifies the application config and dependencies can initialize.
func TestAppStartup(t *testing.T) {
	// TODO: load config, init DB/Redis connections, verify health checks
	t.Skip("integration test — requires running MySQL and Redis")
}
