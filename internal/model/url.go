package model

import (
	"net/url"
	"path"
	"strconv"
	"strings"
)

// ParsedURL holds the decomposed components of a URL.
type ParsedURL struct {
	Scheme   string `json:"scheme"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Host     string `json:"host,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Port     string `json:"port,omitempty"`
	Path     string `json:"path,omitempty"`
	Query    string `json:"query,omitempty"`
	Fragment string `json:"fragment,omitempty"`
	Origin   string `json:"origin,omitempty"`
	// Derived
	Domain     string `json:"domain,omitempty"`
	Subdomain string `json:"subdomain,omitempty"`
	TLD        string `json:"tld,omitempty"`
	IsHTTPS    bool   `json:"is_https"`
	IsSecure   bool   `json:"is_secure"`
	HasPort    bool   `json:"has_port"`
	HasQuery   bool   `json:"has_query"`
	HasFrag    bool   `json:"has_fragment"`
	HasUserInfo bool  `json:"has_user_info,omitempty"`
}

// Parse decomposes a URL string into its components.
func Parse(rawURL string) (*ParsedURL, error) {
	if strings.TrimSpace(rawURL) == "" {
		return nil, ErrEmptyURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		return nil, ErrNoScheme
	}

	p := &ParsedURL{
		Scheme:   u.Scheme,
		Username: u.User.Username(),
		Host:     u.Host,
		Hostname: u.Hostname(),
		Path:     u.Path,
		Query:    u.RawQuery,
		Fragment: u.Fragment,
	}

	if u.User != nil {
		p.Password, _ = u.User.Password()
		p.HasUserInfo = true
	}

	if portStr := u.Port(); portStr != "" {
		p.Port = portStr
		p.HasPort = true
	}

	if u.Scheme == "https" || u.Scheme == "wss" {
		p.IsHTTPS = true
		p.IsSecure = true
	}

	if u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ws" || u.Scheme == "wss" {
		p.Origin = u.Scheme + "://" + u.Host
	}

	if p.Query != "" {
		p.HasQuery = true
	}
	if p.Fragment != "" {
		p.HasFrag = true
	}

	// Extract domain parts from hostname
	if p.Hostname != "" && !isIPAddress(p.Hostname) {
		parts := strings.Split(p.Hostname, ".")
		if len(parts) >= 2 {
			p.Domain = parts[len(parts)-2] + "." + parts[len(parts)-1]
			p.TLD = parts[len(parts)-1]
			if len(parts) > 2 {
				p.Subdomain = strings.Join(parts[:len(parts)-2], ".")
			}
		}
	}

	return p, nil
}

// QueryParam represents a single query parameter key-value pair.
type QueryParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ParseQuery parses a raw query string into a list of key-value pairs.
func ParseQuery(rawQuery string) ([]QueryParam, error) {
	if rawQuery == "" {
		return nil, nil
	}
	vals, err := url.ParseQuery(rawQuery)
	if err != nil {
		return nil, err
	}
	var params []QueryParam
	for key, vSlice := range vals {
		for _, v := range vSlice {
			params = append(params, QueryParam{Key: key, Value: v})
		}
	}
	return params, nil
}

// AddQueryParam adds a key-value pair to an existing raw query string and returns the new query string.
func AddQueryParam(rawQuery, key, value string) string {
	vals, _ := url.ParseQuery(rawQuery)
	vals.Add(key, value)
	return vals.Encode()
}

// RemoveQueryParam removes all occurrences of a key from a raw query string.
func RemoveQueryParam(rawQuery, key string) string {
	vals, _ := url.ParseQuery(rawQuery)
	vals.Del(key)
	return vals.Encode()
}

// SetQueryParam sets a key to a single value, replacing any existing values.
func SetQueryParam(rawQuery, key, value string) string {
	vals, _ := url.ParseQuery(rawQuery)
	vals.Set(key, value)
	return vals.Encode()
}

// GetQueryParam retrieves the first value for a key from a raw query string.
func GetQueryParam(rawQuery, key string) (string, bool) {
	vals, _ := url.ParseQuery(rawQuery)
	v := vals.Get(key)
	return v, v != "" || vals.Has(key)
}

// BuildURL constructs a URL from individual components.
type URLParts struct {
	Scheme   string `json:"scheme"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Host     string `json:"host,omitempty"`
	Path     string `json:"path,omitempty"`
	Query    string `json:"query,omitempty"`
	Fragment string `json:"fragment,omitempty"`
}

// Build constructs a URL string from the given parts.
func (p URLParts) Build() string {
	u := &url.URL{}
	u.Scheme = p.Scheme
	if p.Username != "" {
		if p.Password != "" {
			u.User = url.UserPassword(p.Username, p.Password)
		} else {
			u.User = url.User(p.Username)
		}
	}
	u.Host = p.Host
	if p.Path != "" {
		u.Path = p.Path
	}
	if p.Query != "" {
		u.RawQuery = p.Query
	}
	if p.Fragment != "" {
		u.Fragment = p.Fragment
	}
	return u.String()
}

// Resolve resolves a reference URL against a base URL (like a browser would).
func Resolve(baseURL, refURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(refURL)
	if err != nil {
		return "", err
	}
	resolved := base.ResolveReference(ref)
	return resolved.String(), nil
}

// Normalize applies common URL normalization rules:
// - Lowercase scheme and host
// - Remove default port (80 for http, 443 for https)
// - Remove duplicate slashes in path
// - Remove trailing slash (unless path is "/")
// - Sort query parameters
// - Remove fragment
func Normalize(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		return "", ErrNoScheme
	}

	// Lowercase scheme and host
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Remove default ports
	host := u.Hostname()
	port := u.Port()
	if (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		u.Host = host
	}

	// Remove duplicate slashes
	u.Path = dedupSlashes(u.Path)

	// Remove trailing slash (unless root)
	if len(u.Path) > 1 && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	// Sort query params
	if u.RawQuery != "" {
		vals, _ := url.ParseQuery(u.RawQuery)
		u.RawQuery = vals.Encode()
	}

	// Remove fragment
	u.Fragment = ""

	return u.String(), nil
}

// PercentEncode encodes a string for use in a URL path segment.
func PercentEncode(s string) string {
	return url.PathEscape(s)
}

// PercentDecode decodes a percent-encoded string.
func PercentDecode(s string) (string, error) {
	return url.PathUnescape(s)
}

// QueryEncode encodes a string for use in a URL query parameter value.
func QueryEncode(s string) string {
	return url.QueryEscape(s)
}

// QueryDecode decodes a URL-encoded query parameter value.
func QueryDecode(s string) (string, error) {
	return url.QueryUnescape(s)
}

// JoinPath joins a base URL with path segments.
func JoinPath(baseURL string, segments ...string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	p := u.Path
	for _, seg := range segments {
		p = path.Join(p, seg)
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u.Path = p
	return u.String(), nil
}

// IsAbsolute returns true if the URL has a scheme.
func IsAbsolute(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.IsAbs()
}

// SameOrigin checks if two URLs share the same origin (scheme + host + port).
func SameOrigin(url1, url2 string) (bool, error) {
	u1, err := url.Parse(url1)
	if err != nil {
		return false, err
	}
	u2, err := url.Parse(url2)
	if err != nil {
		return false, err
	}
	return u1.Scheme == u2.Scheme && u1.Host == u2.Host, nil
}

// IsIP checks if a string is an IP address (IPv4 or IPv6).
func isIPAddress(s string) bool {
	// Simple check: IPv4 has dots and all-numeric parts, IPv6 has colons
	if strings.Contains(s, ":") {
		return true // likely IPv6
	}
	parts := strings.Split(s, ".")
	if len(parts) == 4 {
		for _, p := range parts {
			if _, err := strconv.Atoi(p); err != nil {
				return false
			}
		}
		return true
	}
	return false
}

func dedupSlashes(p string) string {
	if p == "" {
		return p
	}
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	return p
}
