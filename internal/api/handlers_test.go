package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/parse?url=https://sub.example.com:8080/api?q=1#top", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "scheme=https") {
		t.Errorf("body missing scheme: %s", body)
	}
	if !strings.Contains(body, "hostname=sub.example.com") {
		t.Errorf("body missing hostname: %s", body)
	}
	if !strings.Contains(body, "domain=example.com") {
		t.Errorf("body missing domain: %s", body)
	}
}

func TestParseJSON(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/parse?url=https://example.com&format=json", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("content-type = %q, want application/json", w.Header().Get("Content-Type"))
	}
}

func TestParseMissingURL(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/parse", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "error:") {
		t.Errorf("body should contain error: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "hint:") {
		t.Errorf("body should contain hint: %s", w.Body.String())
	}
}

func TestBuild(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/build?scheme=https&host=example.com&path=/api&query=a=1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "url=https://example.com/api?a=1") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestResolve(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/resolve?base=https://example.com/a/b/c&ref=../d", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "url=https://example.com/a/d") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestNormalize(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/normalize?url=https%3A%2F%2FExample.com:443%2Fa%2F%2Fb%2F%3Fz%3D2%26a%3D1%23frag", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "url=https://example.com/a/b?a=1&z=2") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestEncode(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/encode?s=hello%20world", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "encoded=hello%20world") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestDecode(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/decode?s=hello%2520world", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "decoded=hello world") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestQueryList(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/query?action=list&q=a%3D1%26b%3D2", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "a=1") || !strings.Contains(body, "b=2") {
		t.Errorf("body = %s", body)
	}
}

func TestQueryAdd(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/query?action=add&q=a%3D1&key=b&value=2", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "query=") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestQueryRemove(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/query?action=remove&q=a%3D1%26b%3D2&key=a", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "query=b=2") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestQueryGet(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/query?action=get&q=a%3D1%26b%3D2&key=a", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "key=a value=1") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestJoin(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/join?base=https://example.com/api&path=v1&path=users", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "url=https://example.com/api/v1/users") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestOrigin(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/origin?url=https://example.com/path", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "origin=https://example.com") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestDomain(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/domain?url=https://sub.example.com/path", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "domain=example.com") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestAbsolute(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/absolute?url=https://example.com", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "absolute=true") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestSameOrigin(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/same-origin?url1=https://a.com&url2=https://a.com/path", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "same_origin=true") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestHelp(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/help", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "urlkit") {
		t.Errorf("body should contain urlkit: %s", w.Body.String())
	}
}

func TestRoot(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "urlkit") {
		t.Errorf("body should contain urlkit: %s", w.Body.String())
	}
}

func TestNotFound(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestMCPInitialize(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize"}`
	req := httptest.NewRequest("POST", "/mcp", io.NopCloser(strings.NewReader(body)))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "urlkit") {
		t.Errorf("body should contain urlkit: %s", w.Body.String())
	}
}

func TestMCPToolsList(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	body := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	req := httptest.NewRequest("POST", "/mcp", io.NopCloser(strings.NewReader(body)))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "parse") {
		t.Errorf("body should contain parse tool: %s", w.Body.String())
	}
}

func TestMCPCallParse(t *testing.T) {
	h := NewHandler()
	mux := h.Routes()

	body := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"parse","arguments":{"url":"https://example.com/path"}}}`
	req := httptest.NewRequest("POST", "/mcp", io.NopCloser(strings.NewReader(body)))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "https") {
		t.Errorf("body should contain https: %s", w.Body.String())
	}
}
