package http

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/opcotech/elemo/internal/pkg/log"
)

func (c *pluginController) ServePluginAsset(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/ServePluginAsset")
	defer span.End()

	pluginID := chi.URLParam(r, "pluginId")
	version := chi.URLParam(r, "version")
	rel := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if pluginID == "" || version == "" || rel == "" {
		http.NotFound(w, r)
		return
	}

	f, err := c.pluginService.OpenAsset(ctx, pluginID, version, rel)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusNotFound, http.StatusBadRequest, http.StatusForbidden:
			http.NotFound(w, r)
		default:
			c.logger.Error(ctx, "failed to serve plugin asset", log.WithError(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			c.logger.Error(ctx, "failed to close plugin asset", log.WithError(cerr))
		}
	}()

	info, err := f.Stat()
	if err != nil {
		c.logger.Error(ctx, "failed to stat plugin asset", log.WithError(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setCommonHeaders(w)
	w.Header().Set("Content-Type", pluginAssetContentType(rel))
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, filepath.Base(rel), info.ModTime(), f)
}

func pluginAssetContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".js", ".mjs":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".wasm":
		return "application/wasm"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".woff2":
		return "font/woff2"
	case ".map":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
