package handler

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Static phục vụ file tĩnh của frontend (dist) + SPA fallback về index.html.
type Static struct {
	dir string
}

// NewStatic tạo handler; dir rỗng → mọi request trả 404 text.
func NewStatic(dir string) *Static { return &Static{dir: dir} }

func (s *Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.dir == "" {
		http.NotFound(w, r)
		return
	}
	clean := path.Clean("/" + r.URL.Path)
	if clean != "/" {
		full := filepath.Join(s.dir, filepath.FromSlash(clean))
		if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
			if strings.HasPrefix(clean, "/assets/") {
				// file có hash trong tên → cache lâu
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "public, max-age=3600")
			}
			http.ServeFile(w, r, full)
			return
		}
		// file tĩnh thiếu (có đuôi) → 404 thật, không trả index.html
		if strings.HasPrefix(clean, "/assets/") {
			http.NotFound(w, r)
			return
		}
	}
	index := filepath.Join(s.dir, "index.html")
	if _, err := os.Stat(index); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, index)
}
