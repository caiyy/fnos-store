package api

import (
	"testing"

	"fnos-store/internal/config"
	"fnos-store/internal/core"
	"fnos-store/internal/source"
)

func newAutoUpdateTestServer(t *testing.T, cfg config.Config) *Server {
	t.Helper()
	mgr := config.NewManager(t.TempDir())
	if err := mgr.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	registry := core.NewRegistry()
	registry.Merge([]core.Manifest{{
		AppName:     "agent2api",
		Version:     "3.5.0-beta.6.6",
		FpkVersion:  "3.5.0-beta.6.6",
		Distributor: "Cocode.Ren",
	}}, []source.RemoteApp{{
		AppName:    "agent2api",
		Version:    "3.5.0-beta.7.1",
		FpkVersion: "3.5.0-beta.7.1",
		ReleaseTag: "fnos-v3.5.0-beta.7.1",
		FpkURL:     "https://example.com/agent2api.fpk",
	}}, nil)

	return &Server{
		registry:  registry,
		configMgr: mgr,
		storeApp:  "cocode-store",
	}
}

func TestAutoUpdateCandidates(t *testing.T) {
	server := newAutoUpdateTestServer(t, config.Config{
		CheckIntervalHours: 1,
		AutoUpdate:         true,
		Mirror:             "gh-proxy",
		DockerMirror:       "daocloud",
	})

	candidates := server.autoUpdateCandidates()
	if len(candidates) != 1 || candidates[0].AppName != "agent2api" {
		t.Fatalf("autoUpdateCandidates() = %+v, want agent2api", candidates)
	}
}

func TestAutoUpdateCandidatesHonorsDisabledAndIgnored(t *testing.T) {
	disabled := newAutoUpdateTestServer(t, config.Config{
		CheckIntervalHours: 1,
		AutoUpdate:         false,
		Mirror:             "gh-proxy",
		DockerMirror:       "daocloud",
	})
	if got := disabled.autoUpdateCandidates(); len(got) != 0 {
		t.Fatalf("disabled auto-update returned %+v", got)
	}

	ignored := newAutoUpdateTestServer(t, config.Config{
		CheckIntervalHours: 1,
		AutoUpdate:         true,
		Mirror:             "gh-proxy",
		DockerMirror:       "daocloud",
		IgnoredApps:        []string{"agent2api"},
	})
	if got := ignored.autoUpdateCandidates(); len(got) != 0 {
		t.Fatalf("ignored app returned %+v", got)
	}
}
