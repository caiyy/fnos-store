package api

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"strings"

	"fnos-store/internal/core"
)

type autoUpdateResponseWriter struct {
	header http.Header
	body   bytes.Buffer
}

func newAutoUpdateResponseWriter() *autoUpdateResponseWriter {
	return &autoUpdateResponseWriter{header: make(http.Header)}
}

func (w *autoUpdateResponseWriter) Header() http.Header {
	return w.header
}

func (w *autoUpdateResponseWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *autoUpdateResponseWriter) WriteHeader(_ int) {}

func (w *autoUpdateResponseWriter) Flush() {}

func (s *Server) autoUpdateCandidates() []core.AppInfo {
	if s.configMgr == nil {
		return nil
	}
	cfg := s.configMgr.Get()
	if !cfg.AutoUpdate {
		return nil
	}

	apps := s.listRegistryApps()
	candidates := make([]core.AppInfo, 0, len(apps))
	for _, app := range apps {
		if app.AppName == s.storeApp || !app.Installed || app.Status != core.AppStatusUpdateAvailable {
			continue
		}
		if cfg.IsAppIgnored(app.AppName) || app.DownloadURL == "" {
			continue
		}
		candidates = append(candidates, app)
	}
	return candidates
}

// CheckAndAutoUpdate is the scheduler entry point. It refreshes the catalog
// first, then sends safe updates through the same data-preserving pipeline used
// by the manual update button.
func (s *Server) CheckAndAutoUpdate(ctx context.Context) error {
	if err := s.refreshRegistry(ctx); err != nil {
		return err
	}
	if s.configMgr == nil || !s.configMgr.Get().AutoUpdate {
		return nil
	}
	if s.ac == nil || s.pipeline == nil {
		return nil
	}
	if capability := s.ac.UpgradeCapability(); !capability.Allowed {
		log.Printf("auto-update skipped: %s", capability.Reason)
		return nil
	}

	for _, app := range s.autoUpdateCandidates() {
		if !s.queue.TryStart("auto-update", app.AppName) {
			log.Printf("auto-update skipped %s: another operation is active", app.AppName)
			continue
		}

		log.Printf("auto-update starting %s %s -> %s", app.AppName, app.InstalledFpkVersion, app.FpkVersion)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/internal/auto-update", nil)
		if err != nil {
			s.queue.FinishApp(app.AppName)
			log.Printf("auto-update failed %s: %v", app.AppName, err)
			continue
		}
		writer := newAutoUpdateResponseWriter()
		stream := &sseStream{
			w:       writer,
			r:       req,
			flusher: writer,
			appname: app.AppName,
		}
		s.pipeline.runStandard(ctx, stream, "update", app, nil, s.refreshRegistry)
		s.queue.FinishApp(app.AppName)

		if body := writer.body.String(); strings.Contains(body, `"step":"error"`) {
			log.Printf("auto-update failed %s: %s", app.AppName, body)
		} else {
			log.Printf("auto-update completed %s -> %s", app.AppName, app.FpkVersion)
		}
	}
	return nil
}
