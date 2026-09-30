package model

import (
	"testing"
)

func TestParse(t *testing.T) {
	p, err := Parse("https://sub.example.com:8080/api/v1?q=1&r=2#top")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Scheme != "https" {
		t.Errorf("scheme = %q, want https", p.Scheme)
	}
	if p.Hostname != "sub.example.com" {
		t.Errorf("hostname = %q, want sub.example.com", p.Hostname)
	}
	if p.Port != "8080" {
		t.Errorf("port = %q, want 8080", p.Port)
	}
	if p.Path != "/api/v1" {
		t.Errorf("path = %q, want /api/v1", p.Path)
	}
	if p.Query != "q=1&r=2" {
		t.Errorf("query = %q, want q=1&r=2", p.Query)
	}
	if p.Fragment != "top" {
		t.Errorf("fragment = %q, want top", p.Fragment)
	}
	if p.Domain != "example.com" {
		t.Errorf("domain = %q, want example.com", p.Domain)
	}
	if p.Subdomain != "sub" {
		t.Errorf("subdomain = %q, want sub", p.Subdomain)
	}
	if p.TLD != "com" {
		t.Errorf("tld = %q, want com", p.TLD)
	}
	if !p.IsHTTPS {
		t.Error("is_https should be true")
	}
	if !p.IsSecure {
		t.Error("is_secure should be true")
	}
	if !p.HasPort {
		t.Error("has_port should be true")
	}
	if !p.HasQuery {
		t.Error("has_query should be true")
	}
	if !p.HasFrag {
		t.Error("has_fragment should be true")
	}
	if p.Origin != "https://sub.example.com:8080" {
		t.Errorf("origin = %q, want https://sub.example.com:8080", p.Origin)
	}
}

func TestParseNoScheme(t *testing.T) {
	_, err := Parse("example.com/path")
	if err != ErrNoScheme {
		t.Errorf("expected ErrNoScheme, got %v", err)
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	if err != ErrEmptyURL {
		t.Errorf("expected ErrEmptyURL, got %v", err)
	}
}

func TestParseUserInfo(t *testing.T) {
	p, err := Parse("https://user:pass@example.com/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Username != "user" {
		t.Errorf("username = %q, want user", p.Username)
	}
	if p.Password != "pass" {
		t.Errorf("password = %q, want pass", p.Password)
	}
	if !p.HasUserInfo {
		t.Error("has_user_info should be true")
	}
}

func TestParseHTTPNotSecure(t *testing.T) {
	p, err := Parse("http://example.com/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.IsHTTPS {
		t.Error("is_https should be false for http")
	}
	if p.IsSecure {
		t.Error("is_secure should be false for http")
	}
}

func TestParseQuery(t *testing.T) {
	params, err := ParseQuery("a=1&b=2&a=3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(params))
	}
}

func TestParseQueryEmpty(t *testing.T) {
	params, err := ParseQuery("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(params) != 0 {
		t.Errorf("expected 0 params, got %d", len(params))
	}
}

func TestAddQueryParam(t *testing.T) {
	result := AddQueryParam("a=1", "b", "2")
	if result != "a=1&b=2" {
		t.Errorf("got %q, want a=1&b=2", result)
	}
}

func TestRemoveQueryParam(t *testing.T) {
	result := RemoveQueryParam("a=1&b=2", "a")
	if result != "b=2" {
		t.Errorf("got %q, want b=2", result)
	}
}

func TestSetQueryParam(t *testing.T) {
	result := SetQueryParam("a=1&b=2", "a", "3")
	if result != "a=3&b=2" {
		t.Errorf("got %q, want a=3&b=2", result)
	}
}

func TestGetQueryParam(t *testing.T) {
	val, ok := GetQueryParam("a=1&b=2", "a")
	if !ok {
		t.Error("expected ok=true")
	}
	if val != "1" {
		t.Errorf("got %q, want 1", val)
	}
}

func TestGetQueryParamMissing(t *testing.T) {
	_, ok := GetQueryParam("a=1", "b")
	if ok {
		t.Error("expected ok=false for missing key")
	}
}

func TestBuild(t *testing.T) {
	parts := URLParts{
		Scheme: "https",
		Host:   "example.com",
		Path:   "/api",
		Query:  "a=1",
	}
	result := parts.Build()
	if result != "https://example.com/api?a=1" {
		t.Errorf("got %q, want https://example.com/api?a=1", result)
	}
}

func TestBuildWithUserInfo(t *testing.T) {
	parts := URLParts{
		Scheme:   "https",
		Username: "user",
		Password: "pass",
		Host:     "example.com",
	}
	result := parts.Build()
	if result != "https://user:pass@example.com" {
		t.Errorf("got %q, want https://user:pass@example.com", result)
	}
}

func TestResolve(t *testing.T) {
	result, err := Resolve("https://example.com/a/b/c", "../d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "https://example.com/a/d" {
		t.Errorf("got %q, want https://example.com/a/d", result)
	}
}

func TestResolveAbsolute(t *testing.T) {
	result, err := Resolve("https://example.com/a/b", "https://other.com/c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "https://other.com/c" {
		t.Errorf("got %q, want https://other.com/c", result)
	}
}

func TestNormalize(t *testing.T) {
	result, err := Normalize("https://Example.com:443/a//b/?z=2&a=1#frag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "https://example.com/a/b?a=1&z=2" {
		t.Errorf("got %q, want https://example.com/a/b?a=1&z=2", result)
	}
}

func TestNormalizeDefaultPort(t *testing.T) {
	result, err := Normalize("http://example.com:80/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "http://example.com/path" {
		t.Errorf("got %q, want http://example.com/path", result)
	}
}

func TestPercentEncode(t *testing.T) {
	result := PercentEncode("hello world")
	if result != "hello%20world" {
		t.Errorf("got %q, want hello%%20world", result)
	}
}

func TestPercentDecode(t *testing.T) {
	result, err := PercentDecode("hello%20world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("got %q, want hello world", result)
	}
}

func TestQueryEncode(t *testing.T) {
	result := QueryEncode("a b&c")
	if result != "a+b%26c" {
		t.Errorf("got %q, want a+b%%26c", result)
	}
}

func TestQueryDecode(t *testing.T) {
	result, err := QueryDecode("a+b%26c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "a b&c" {
		t.Errorf("got %q, want a b&c", result)
	}
}

func TestJoinPath(t *testing.T) {
	result, err := JoinPath("https://example.com/api", "v1", "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "https://example.com/api/v1/users" {
		t.Errorf("got %q, want https://example.com/api/v1/users", result)
	}
}

func TestIsAbsolute(t *testing.T) {
	if !IsAbsolute("https://example.com") {
		t.Error("expected true for absolute URL")
	}
	if IsAbsolute("/relative/path") {
		t.Error("expected false for relative URL")
	}
}

func TestSameOrigin(t *testing.T) {
	result, err := SameOrigin("https://example.com/a", "https://example.com/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected same origin")
	}
}

func TestSameOriginDifferent(t *testing.T) {
	result, err := SameOrigin("https://a.com", "https://b.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result {
		t.Error("expected different origin")
	}
}

func TestDedupSlashes(t *testing.T) {
	if dedupSlashes("/a//b///c") != "/a/b/c" {
		t.Error("dedupSlashes failed")
	}
}
