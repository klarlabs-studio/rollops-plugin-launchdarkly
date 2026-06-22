// Package launchdarkly is a Rollops feature-flag provider plugin backed by
// LaunchDarkly's REST API. It drives a flag's on/off state and its
// environment-level percentage rollout (fallthrough) to match a rollout's
// progressive steps, so a LaunchDarkly flag tracks a Rollops canary in lockstep.
package launchdarkly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.klarlabs.de/rollops/pkg/plugin"
)

// Provider talks to LaunchDarkly's REST API. BaseURL, Token, and Project come
// from the plugin's environment (see Config); Environment is supplied per call
// by Rollops as the LaunchDarkly environment key (e.g. "production"). The flag's
// first two variations are assumed to be the true/false pair (the LaunchDarkly
// default for a boolean flag): variation 0 = on, variation 1 = off.
type Provider struct {
	BaseURL string // e.g. https://app.launchdarkly.com
	Token   string // API access token (Authorization: <token>)
	Project string // project key (default "default")
	HTTP    *http.Client
}

func (p Provider) client() *http.Client {
	if p.HTTP != nil {
		return p.HTTP
	}
	return http.DefaultClient
}

func (p Provider) project() string {
	if p.Project != "" {
		return p.Project
	}
	return "default"
}

// ApplyFlag PATCHes the flag with a JSON Patch that sets the environment's on
// state and replaces its fallthrough with a percentage rollout. LaunchDarkly
// weights are per-mille (0–100000), so a percentage is scaled by 1000.
func (p Provider) ApplyFlag(ctx context.Context, c plugin.FlagChange) error {
	if p.Token == "" {
		return fmt.Errorf("launchdarkly: LAUNCHDARKLY_TOKEN is required")
	}
	weight := c.Percentage * 1000
	rollout := map[string]any{
		"rollout": map[string]any{
			"variations": []any{
				map[string]any{"variation": 0, "weight": weight},
				map[string]any{"variation": 1, "weight": 100000 - weight},
			},
		},
	}
	patch := []any{
		map[string]any{"op": "replace", "path": fmt.Sprintf("/environments/%s/on", c.Environment), "value": !c.Disabled},
		map[string]any{"op": "replace", "path": fmt.Sprintf("/environments/%s/fallthrough", c.Environment), "value": rollout},
	}
	u := fmt.Sprintf("%s/api/v2/flags/%s/%s", p.BaseURL, url.PathEscape(p.project()), url.PathEscape(c.Flag))
	if err := p.patch(ctx, u, patch); err != nil {
		return fmt.Errorf("launchdarkly: update flag %q: %w", c.Flag, err)
	}
	return nil
}

func (p Provider) patch(ctx context.Context, u string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", p.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client().Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
