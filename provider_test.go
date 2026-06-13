package launchdarkly

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.klarlabs.de/rollops/pkg/plugin"
)

func TestApplyFlag_PatchesRolloutAndOn(t *testing.T) {
	var path string
	var patch []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &patch)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := Provider{BaseURL: srv.URL, Token: "tok", Project: "default", HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "checkout", Environment: "production", Percentage: 25}); err != nil {
		t.Fatalf("ApplyFlag: %v", err)
	}
	if !strings.HasSuffix(path, "/api/v2/flags/default/checkout") {
		t.Errorf("wrong path: %s", path)
	}
	if patch[0]["path"] != "/environments/production/on" || patch[0]["value"] != true {
		t.Errorf("on op = %v, want on=true", patch[0])
	}
	val, _ := patch[1]["value"].(map[string]any)
	rollout, _ := val["rollout"].(map[string]any)
	vars, _ := rollout["variations"].([]any)
	v0, _ := vars[0].(map[string]any)
	v1, _ := vars[1].(map[string]any)
	if v0["weight"].(float64) != 25000 || v1["weight"].(float64) != 75000 {
		t.Errorf("weights = %v/%v, want 25000/75000 per-mille", v0["weight"], v1["weight"])
	}
}

func TestApplyFlag_DisabledTurnsOff(t *testing.T) {
	var patch []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &patch)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := Provider{BaseURL: srv.URL, Token: "tok", HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f", Environment: "production", Disabled: true}); err != nil {
		t.Fatalf("ApplyFlag: %v", err)
	}
	if patch[0]["value"] != false {
		t.Errorf("on op value = %v, want false", patch[0]["value"])
	}
}

func TestApplyFlag_ServerErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(401) }))
	defer srv.Close()
	p := Provider{BaseURL: srv.URL, Token: "tok", HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f", Environment: "p"}); err == nil {
		t.Fatal("401 must error")
	}
}

func TestApplyFlag_RequiresToken(t *testing.T) {
	p := Provider{BaseURL: "http://x"}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f", Environment: "p"}); err == nil {
		t.Fatal("missing token must error")
	}
}
