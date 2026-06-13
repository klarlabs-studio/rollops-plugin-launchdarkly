// Command rollops-plugin-launchdarkly is a Rollops feature-flag provider plugin
// backed by LaunchDarkly. Build it, pin its sha256, and point a rollout's
// featureFlags.plugin at the binary.
package main

import (
	"fmt"
	"os"

	launchdarkly "github.com/klarlabs-studio/rollops-plugin-launchdarkly"
	"go.klarlabs.de/rollops/pkg/plugin"
)

// version is overwritten at build time via -ldflags.
var version = "dev"

func main() {
	safety := plugin.Safety{
		NetworkHosts: []string{"app.launchdarkly.com:443"},
		EnvVars:      []string{"LAUNCHDARKLY_API_URL", "LAUNCHDARKLY_TOKEN", "LAUNCHDARKLY_PROJECT"},
		RiskClass:    plugin.RiskActive,
	}
	if err := plugin.ServeFlagProvider("klarlabs/launchdarkly", version, launchdarkly.FromEnv(), safety); err != nil {
		fmt.Fprintln(os.Stderr, "rollops-plugin-launchdarkly:", err)
		os.Exit(1)
	}
}
