package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/relentlessworks/urlkit/internal/model"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/help", h.help)
	mux.HandleFunc("/.well-known/agent.md", h.help)
	mux.HandleFunc("/parse", h.parse)
	mux.HandleFunc("/build", h.build)
	mux.HandleFunc("/resolve", h.resolve)
	mux.HandleFunc("/normalize", h.normalize)
	mux.HandleFunc("/encode", h.encode)
	mux.HandleFunc("/decode", h.decode)
	mux.HandleFunc("/query", h.query)
	mux.HandleFunc("/join", h.join)
	mux.HandleFunc("/origin", h.origin)
	mux.HandleFunc("/domain", h.domain)
	mux.HandleFunc("/absolute", h.absolute)
	mux.HandleFunc("/same-origin", h.sameOrigin)
	mux.HandleFunc("/mcp", h.mcp)
	mux.HandleFunc("/", h.root)
	return mux
}

func wantsJSON(r *http.Request) bool {
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}

func writeError(w http.ResponseWriter, status int, msg, hint string) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "error: %s | hint: %s\n", msg, hint)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		fmt.Fprintln(w, "urlkit — agentic-first URL parsing and manipulation service")
		fmt.Fprintln(w, "GET /help for the operating manual")
		return
	}
	writeError(w, http.StatusNotFound, "unknown endpoint: "+r.URL.Path, "GET /help to see available endpoints")
}

func (h *Handler) help(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, helpText)
}

func (h *Handler) parse(w http.ResponseWriter, r *http.Request) {
	rawURL := getURLParam(r, "url")
	if rawURL == "" {
		writeError(w, http.StatusBadRequest, "missing 'url' parameter", "GET /parse?url=https://example.com/path?q=1")
		return
	}
	p, err := model.Parse(rawURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "provide a full URL with scheme, e.g. https://example.com/path")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, p)
		return
	}
	// Plain text output
	fmt.Fprintf(w, "scheme=%s host=%s hostname=%s port=%s path=%s query=%s fragment=%s\n",
		p.Scheme, p.Host, p.Hostname, p.Port, p.Path, p.Query, p.Fragment)
	if p.Username != "" {
		fmt.Fprintf(w, "username=%s password=%s\n", p.Username, p.Password)
	}
	if p.Domain != "" {
		fmt.Fprintf(w, "domain=%s subdomain=%s tld=%s\n", p.Domain, p.Subdomain, p.TLD)
	}
	if p.Origin != "" {
		fmt.Fprintf(w, "origin=%s\n", p.Origin)
	}
	fmt.Fprintf(w, "is_https=%v is_secure=%v has_port=%v has_query=%v has_fragment=%v\n",
		p.IsHTTPS, p.IsSecure, p.HasPort, p.HasQuery, p.HasFrag)
}

func (h *Handler) build(w http.ResponseWriter, r *http.Request) {
	parts := model.URLParts{
		Scheme:   r.URL.Query().Get("scheme"),
		Username: r.URL.Query().Get("username"),
		Password: r.URL.Query().Get("password"),
		Host:     r.URL.Query().Get("host"),
		Path:     r.URL.Query().Get("path"),
		Query:    r.URL.Query().Get("query"),
		Fragment: r.URL.Query().Get("fragment"),
	}
	if parts.Scheme == "" {
		writeError(w, http.StatusBadRequest, "missing 'scheme' parameter", "GET /build?scheme=https&host=example.com&path=/api")
		return
	}
	result := parts.Build()
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]string{"url": result})
		return
	}
	fmt.Fprintf(w, "url=%s\n", result)
}

func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	ref := r.URL.Query().Get("ref")
	if base == "" || ref == "" {
		writeError(w, http.StatusBadRequest, "missing 'base' or 'ref' parameter", "GET /resolve?base=https://example.com/a/b&ref=../c")
		return
	}
	result, err := model.Resolve(base, ref)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "provide valid base and ref URLs")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]string{"url": result})
		return
	}
	fmt.Fprintf(w, "url=%s\n", result)
}

func (h *Handler) normalize(w http.ResponseWriter, r *http.Request) {
	rawURL := getURLParam(r, "url")
	if rawURL == "" {
		writeError(w, http.StatusBadRequest, "missing 'url' parameter", "GET /normalize?url=https://Example.com:443/a//b/../c/?z=2&a=1#frag")
		return
	}
	result, err := model.Normalize(rawURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "provide a full URL with scheme")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]string{"url": result})
		return
	}
	fmt.Fprintf(w, "url=%s\n", result)
}

func (h *Handler) encode(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("s")
	if s == "" {
		writeError(w, http.StatusBadRequest, "missing 's' parameter", "GET /encode?s=hello world")
		return
	}
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "path"
	}
	var result string
	switch mode {
	case "path":
		result = model.PercentEncode(s)
	case "query":
		result = model.QueryEncode(s)
	default:
		writeError(w, http.StatusBadRequest, "unknown mode: "+mode, "use mode=path or mode=query")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]string{"encoded": result})
		return
	}
	fmt.Fprintf(w, "encoded=%s\n", result)
}

func (h *Handler) decode(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("s")
	if s == "" {
		writeError(w, http.StatusBadRequest, "missing 's' parameter", "GET /decode?s=hello%20world")
		return
	}
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "path"
	}
	var result string
	var err error
	switch mode {
	case "path":
		result, err = model.PercentDecode(s)
	case "query":
		result, err = model.QueryDecode(s)
	default:
		writeError(w, http.StatusBadRequest, "unknown mode: "+mode, "use mode=path or mode=query")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "provide a valid percent-encoded string")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]string{"decoded": result})
		return
	}
	fmt.Fprintf(w, "decoded=%s\n", result)
}

func (h *Handler) query(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	if action == "" {
		action = "list"
	}
	rawQuery := r.URL.Query().Get("q")
	switch action {
	case "list":
		if rawQuery == "" {
			writeError(w, http.StatusBadRequest, "missing 'q' parameter", "GET /query?action=list&q=a=1&b=2")
			return
		}
		params, err := model.ParseQuery(rawQuery)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), "provide a valid query string")
			return
		}
		if wantsJSON(r) {
			writeJSON(w, http.StatusOK, params)
			return
		}
		if len(params) == 0 {
			fmt.Fprintln(w, "(no parameters)")
			return
		}
		for _, p := range params {
			fmt.Fprintf(w, "%s=%s\n", p.Key, p.Value)
		}
	case "add":
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("value")
		if key == "" {
			writeError(w, http.StatusBadRequest, "missing 'key' parameter", "GET /query?action=add&q=a=1&key=b&value=2")
			return
		}
		result := model.AddQueryParam(rawQuery, key, val)
		if wantsJSON(r) {
			writeJSON(w, http.StatusOK, map[string]string{"query": result})
			return
		}
		fmt.Fprintf(w, "query=%s\n", result)
	case "remove":
		key := r.URL.Query().Get("key")
		if key == "" {
			writeError(w, http.StatusBadRequest, "missing 'key' parameter", "GET /query?action=remove&q=a=1&b=2&key=a")
			return
		}
		result := model.RemoveQueryParam(rawQuery, key)
		if wantsJSON(r) {
			writeJSON(w, http.StatusOK, map[string]string{"query": result})
			return
		}
		fmt.Fprintf(w, "query=%s\n", result)
	case "set":
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("value")
		if key == "" {
			writeError(w, http.StatusBadRequest, "missing 'key' parameter", "GET /query?action=set&q=a=1&b=2&key=a&value=3")
			return
		}
		result := model.SetQueryParam(rawQuery, key, val)
		if wantsJSON(r) {
			writeJSON(w, http.StatusOK, map[string]string{"query": result})
			return
		}
		fmt.Fprintf(w, "query=%s\n", result)
	case "get":
		key := r.URL.Query().Get("key")
		if key == "" {
			writeError(w, http.StatusBadRequest, "missing 'key' parameter", "GET /query?action=get&q=a=1&b=2&key=a")
			return
		}
		val, ok := model.GetQueryParam(rawQuery, key)
		if !ok {
			writeError(w, http.StatusNotFound, "key not found: "+key, "check the query string or use action=list to see all keys")
			return
		}
		if wantsJSON(r) {
			writeJSON(w, http.StatusOK, map[string]string{"key": key, "value": val})
			return
		}
		fmt.Fprintf(w, "key=%s value=%s\n", key, val)
	default:
		writeError(w, http.StatusBadRequest, "unknown action: "+action, "use action=list, add, remove, set, or get")
	}
}

func (h *Handler) join(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	if base == "" {
		writeError(w, http.StatusBadRequest, "missing 'base' parameter", "GET /join?base=https://example.com/api&path=v1&path=users")
		return
	}
	paths := r.URL.Query()["path"]
	result, err := model.JoinPath(base, paths...)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "provide a valid base URL")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]string{"url": result})
		return
	}
	fmt.Fprintf(w, "url=%s\n", result)
}

func (h *Handler) origin(w http.ResponseWriter, r *http.Request) {
	rawURL := getURLParam(r, "url")
	if rawURL == "" {
		writeError(w, http.StatusBadRequest, "missing 'url' parameter", "GET /origin?url=https://example.com/path")
		return
	}
	p, err := model.Parse(rawURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "provide a full URL with scheme")
		return
	}
	if p.Origin == "" {
		writeError(w, http.StatusBadRequest, "no origin for scheme: "+p.Scheme, "origin is only available for http/https/ws/wss URLs")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]string{"origin": p.Origin})
		return
	}
	fmt.Fprintf(w, "origin=%s\n", p.Origin)
}

func (h *Handler) domain(w http.ResponseWriter, r *http.Request) {
	rawURL := getURLParam(r, "url")
	if rawURL == "" {
		writeError(w, http.StatusBadRequest, "missing 'url' parameter", "GET /domain?url=https://sub.example.com/path")
		return
	}
	p, err := model.Parse(rawURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "provide a full URL with scheme")
		return
	}
	if p.Domain == "" {
		writeError(w, http.StatusBadRequest, "no domain found in hostname: "+p.Hostname, "the hostname may be an IP address or a single-label host")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]string{"domain": p.Domain, "subdomain": p.Subdomain, "tld": p.TLD})
		return
	}
	fmt.Fprintf(w, "domain=%s subdomain=%s tld=%s\n", p.Domain, p.Subdomain, p.TLD)
}

func (h *Handler) absolute(w http.ResponseWriter, r *http.Request) {
	rawURL := getURLParam(r, "url")
	if rawURL == "" {
		writeError(w, http.StatusBadRequest, "missing 'url' parameter", "GET /absolute?url=https://example.com/path")
		return
	}
	result := model.IsAbsolute(rawURL)
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]bool{"absolute": result})
		return
	}
	fmt.Fprintf(w, "absolute=%v\n", result)
}

func (h *Handler) sameOrigin(w http.ResponseWriter, r *http.Request) {
	url1 := r.URL.Query().Get("url1")
	url2 := r.URL.Query().Get("url2")
	if url1 == "" || url2 == "" {
		writeError(w, http.StatusBadRequest, "missing 'url1' or 'url2' parameter", "GET /same-origin?url1=https://a.com&url2=https://a.com/path")
		return
	}
	result, err := model.SameOrigin(url1, url2)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "provide valid URLs")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]bool{"same_origin": result})
		return
	}
	fmt.Fprintf(w, "same_origin=%v\n", result)
}

// getURLParam extracts the URL from query params, handling the case where
// the URL itself contains query params that shouldn't be parsed by the server.
func getURLParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}
