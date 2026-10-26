package kitsune

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAudit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!doctype html><html><head></head><body><p>No title here.</p></body></html>`))
	}))
	defer srv.Close()

	r, err := Audit(srv.URL+"/", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != 200 || len(r.Findings) == 0 {
		t.Fatalf("report = %+v", r)
	}
	seen := map[string]bool{}
	for _, f := range r.Findings {
		if f.ID == "" || f.Title == "" {
			t.Errorf("finding = %+v", f)
		}
		seen[f.Severity] = true
	}
	if !seen["error"] {
		t.Errorf("a page without a title has no error finding: %+v", r.Findings)
	}
	if _, err := Audit("http://127.0.0.1:1/", time.Second); err == nil {
		t.Error("no error for a site that isn't there")
	}
}
