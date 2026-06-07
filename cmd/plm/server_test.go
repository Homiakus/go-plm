package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocalAPIMiddlewareRequiresToken(t *testing.T) {
	srv := &Server{apiToken: "secret"}
	handler := srv.localAPIMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/api", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	req = httptest.NewRequest(http.MethodPost, "/api", nil)
	req.AddCookie(&http.Cookie{Name: apiTokenCookie, Value: "secret"})
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status with token = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestLocalAPIMiddlewareRejectsForeignOrigin(t *testing.T) {
	srv := &Server{apiToken: "secret"}
	handler := srv.localAPIMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/api", nil)
	req.Host = "127.0.0.1:8470"
	req.Header.Set("Origin", "https://example.com")
	req.AddCookie(&http.Cookie{Name: apiTokenCookie, Value: "secret"})
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestLocalAPIMiddlewareRequiresSameHostOrigin(t *testing.T) {
	srv := &Server{apiToken: "secret"}
	handler := srv.localAPIMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/api", nil)
	req.Host = "127.0.0.1:8470"
	req.Header.Set("Origin", "http://127.0.0.1:8470")
	req.AddCookie(&http.Cookie{Name: apiTokenCookie, Value: "secret"})
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("same-host status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	req = httptest.NewRequest(http.MethodPost, "/api", nil)
	req.Host = "127.0.0.1:8470"
	req.Header.Set("Origin", "http://127.0.0.1:9999")
	req.AddCookie(&http.Cookie{Name: apiTokenCookie, Value: "secret"})
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-port status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestFrontendHandlerSetsAPITokenCookie(t *testing.T) {
	srv := &Server{apiToken: "secret"}
	handler := srv.frontendHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	if cookies[0].Name != apiTokenCookie || cookies[0].Value != "secret" || !cookies[0].HttpOnly {
		t.Fatalf("unexpected cookie: %#v", cookies[0])
	}
}
