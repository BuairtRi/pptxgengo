// Command wmds-docs serves the West Monroe design-system documentation site: the reference board (foundations,
// primitives, components, composites, frames, templates and catalog) built by tools/build_docs.py.
//
// The same program lives in wm-design-system as docs-server/. It uses only the standard library.
//
//	go run ./cmd/wmdsdocs                      # serves wmds-docs/site on http://localhost:8787
//	go run ./cmd/wmdsdocs -addr :8787           # share on your network
//
// Routes: / (the board), /assets/*, /catalog.json, /changes.json, /CHANGELOG.md, /web-CHANGELOG.md,
// /api/source (SOURCE.json: commit and counts), /healthz.
package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var version = "dev"
var releaseIdentity = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--build-info" {
		fmt.Println(releaseIdentity)
		return
	}
	dir := flag.String("dir", defaultDir(), "folder with the built docs site (index.html, assets/, SOURCE.json)")
	addr := flag.String("addr", "localhost:8787", "listen address; use :8787 to share on your network")
	flag.Parse()

	abs, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(abs, "index.html")); err != nil {
		log.Fatalf("no index.html in %s: build the site first (python3 tools/build_docs.py)", abs)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ok\n") })
	mux.HandleFunc("/api/source", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join(abs, "SOURCE.json"))
	})
	files := http.FileServer(http.Dir(abs))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".md") {
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/assets/"):
			// assets are content-hashed, so they can be cached for good
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		default:
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	}))

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("WM design system docs: %s\n  serving %s%s\n", url(ln.Addr()), abs, describe(abs))
	srv := &http.Server{Handler: logRequests(gzipped(mux)), ReadHeaderTimeout: 10 * time.Second}
	log.Fatal(srv.Serve(ln))
}

// defaultDir finds build/docs-site (design repo) or wmds-docs/site (pptxgengo) from the working directory upward.
func defaultDir() string {
	wd, _ := os.Getwd()
	for d := wd; ; d = filepath.Dir(d) {
		for _, c := range []string{"build/docs-site", "wmds-docs/site"} {
			if _, err := os.Stat(filepath.Join(d, c, "index.html")); err == nil {
				return filepath.Join(d, c)
			}
		}
		if filepath.Dir(d) == d {
			return "build/docs-site"
		}
	}
}

func describe(dir string) string {
	b, err := os.ReadFile(filepath.Join(dir, "SOURCE.json"))
	if err != nil {
		return ""
	}
	var s struct {
		Short     string `json:"short"`
		Dirty     bool   `json:"dirty"`
		Built     string `json:"built"`
		Templates int    `json:"templates"`
		Families  int    `json:"families"`
	}
	if json.Unmarshal(b, &s) != nil {
		return ""
	}
	dirty := ""
	if s.Dirty {
		dirty = "+ (uncommitted)"
	}
	return fmt.Sprintf("\n  wm-design-system %s%s · %d templates in %d families · built %s", s.Short, dirty, s.Templates, s.Families, s.Built)
}

func url(a net.Addr) string {
	host, port, _ := net.SplitHostPort(a.String())
	if host == "" || host == "::" || host == "0.0.0.0" {
		host = "localhost"
	}
	return "http://" + net.JoinHostPort(host, port) + "/"
}

type gzipWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func (g gzipWriter) Write(b []byte) (int, error) { return g.zw.Write(b) }

// gzipped compresses text responses (the board's index is several MB of HTML and JSON).
func gzipped(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		text := p == "/" || strings.HasSuffix(p, ".html") || strings.HasSuffix(p, ".json") || strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".svg")
		if !text || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") || r.Method == http.MethodHead {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
		zw := gzip.NewWriter(w)
		defer zw.Close()
		h.ServeHTTP(gzipWriter{ResponseWriter: w, zw: zw}, stripRange(r))
	})
}

// stripRange drops Range headers: ranges over compressed bodies would be wrong.
func stripRange(r *http.Request) *http.Request {
	if r.Header.Get("Range") == "" {
		return r
	}
	r2 := r.Clone(r.Context())
	r2.Header.Del("Range")
	return r2
}

func logRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
