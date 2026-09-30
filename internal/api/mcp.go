package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/urlkit/internal/model"
)

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *mcpError   `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema struct {
		Type       string         `json:"type"`
		Properties map[string]any `json:"properties"`
		Required   []string       `json:"required,omitempty"`
	} `json:"inputSchema"`
}

type mcpCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type mcpCallResult struct {
	Content []mcpContent `json:"content"`
}

func (h *Handler) mcp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required", "send a JSON-RPC 2.0 POST request to /mcp")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot read body", "send a valid JSON-RPC 2.0 request")
		return
	}

	var req mcpRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON", "send a valid JSON-RPC 2.0 request")
		return
	}

	if req.JSONRPC != "2.0" {
		writeError(w, http.StatusBadRequest, "jsonrpc must be 2.0", "set jsonrpc to \"2.0\"")
		return
	}

	switch req.Method {
	case "initialize":
		h.mcpResult(w, req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"serverInfo": map[string]any{
				"name":    "urlkit",
				"version": "0.1.0",
			},
		})

	case "tools/list":
		h.mcpResult(w, req.ID, map[string]any{
			"tools": mcpTools(),
		})

	case "tools/call":
		var params mcpCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			h.mcpErr(w, req.ID, -32602, "invalid params")
			return
		}
		result, err := h.mcpCallTool(params)
		if err != nil {
			h.mcpErr(w, req.ID, -32603, err.Error())
			return
		}
		h.mcpResult(w, req.ID, result)

	default:
		h.mcpErr(w, req.ID, -32601, "method not found: "+req.Method)
	}
}

func (h *Handler) mcpResult(w http.ResponseWriter, id interface{}, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mcpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func (h *Handler) mcpErr(w http.ResponseWriter, id interface{}, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mcpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &mcpError{Code: code, Message: msg},
	})
}

func mcpTools() []mcpTool {
	str := func(desc string) map[string]any {
		return map[string]any{"type": "string", "description": desc}
	}
	mk := func(name, desc string, props map[string]any, req []string) mcpTool {
		t := mcpTool{Name: name, Description: desc}
		t.InputSchema.Type = "object"
		t.InputSchema.Properties = props
		t.InputSchema.Required = req
		return t
	}
	return []mcpTool{
		mk("parse", "Parse a URL into its components", map[string]any{"url": str("The URL to parse")}, []string{"url"}),
		mk("build", "Build a URL from components", map[string]any{
			"scheme": str("URL scheme"), "host": str("Host with optional port"),
			"path": str("URL path"), "query": str("Query string"), "fragment": str("Fragment"),
		}, []string{"scheme"}),
		mk("resolve", "Resolve a relative URL against a base URL", map[string]any{
			"base": str("Base URL"), "ref": str("Reference URL to resolve"),
		}, []string{"base", "ref"}),
		mk("normalize", "Normalize a URL", map[string]any{"url": str("URL to normalize")}, []string{"url"}),
		mk("encode", "Percent-encode a string", map[string]any{
			"s": str("String to encode"), "mode": str("path or query (default: path)"),
		}, []string{"s"}),
		mk("decode", "Percent-decode a string", map[string]any{
			"s": str("String to decode"), "mode": str("path or query (default: path)"),
		}, []string{"s"}),
		mk("query_list", "List query parameters", map[string]any{"q": str("Raw query string")}, []string{"q"}),
		mk("query_add", "Add a query parameter", map[string]any{
			"q": str("Raw query string"), "key": str("Parameter key"), "value": str("Parameter value"),
		}, []string{"q", "key"}),
		mk("query_remove", "Remove a query parameter", map[string]any{
			"q": str("Raw query string"), "key": str("Parameter key to remove"),
		}, []string{"q", "key"}),
		mk("query_set", "Set a query parameter", map[string]any{
			"q": str("Raw query string"), "key": str("Parameter key"), "value": str("Parameter value"),
		}, []string{"q", "key"}),
		mk("query_get", "Get a query parameter value", map[string]any{
			"q": str("Raw query string"), "key": str("Parameter key"),
		}, []string{"q", "key"}),
		mk("join", "Join path segments to a base URL", map[string]any{
			"base": str("Base URL"), "paths": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Path segments"},
		}, []string{"base"}),
		mk("origin", "Extract origin from URL", map[string]any{"url": str("URL")}, []string{"url"}),
		mk("domain", "Extract domain, subdomain, TLD from URL", map[string]any{"url": str("URL")}, []string{"url"}),
		mk("absolute", "Check if URL is absolute", map[string]any{"url": str("URL")}, []string{"url"}),
		mk("same_origin", "Check if two URLs share the same origin", map[string]any{
			"url1": str("First URL"), "url2": str("Second URL"),
		}, []string{"url1", "url2"}),
	}
}

func (h *Handler) mcpCallTool(params mcpCallParams) (*mcpCallResult, error) {
	getStr := func(key string) string {
		if v, ok := params.Arguments[key].(string); ok {
			return v
		}
		return ""
	}

	text := func(s string) *mcpCallResult {
		return &mcpCallResult{Content: []mcpContent{{Type: "text", Text: s}}}
	}

	switch params.Name {
	case "parse":
		p, err := model.Parse(getStr("url"))
		if err != nil {
			return nil, err
		}
		b, _ := json.Marshal(p)
		return text(string(b)), nil

	case "build":
		parts := model.URLParts{
			Scheme:   getStr("scheme"),
			Host:     getStr("host"),
			Path:     getStr("path"),
			Query:    getStr("query"),
			Fragment: getStr("fragment"),
			Username: getStr("username"),
			Password: getStr("password"),
		}
		return text(fmt.Sprintf("url=%s", parts.Build())), nil

	case "resolve":
		result, err := model.Resolve(getStr("base"), getStr("ref"))
		if err != nil {
			return nil, err
		}
		return text(fmt.Sprintf("url=%s", result)), nil

	case "normalize":
		result, err := model.Normalize(getStr("url"))
		if err != nil {
			return nil, err
		}
		return text(fmt.Sprintf("url=%s", result)), nil

	case "encode":
		mode := getStr("mode")
		if mode == "" {
			mode = "path"
		}
		var result string
		if mode == "query" {
			result = model.QueryEncode(getStr("s"))
		} else {
			result = model.PercentEncode(getStr("s"))
		}
		return text(fmt.Sprintf("encoded=%s", result)), nil

	case "decode":
		mode := getStr("mode")
		if mode == "" {
			mode = "path"
		}
		var result string
		var derr error
		if mode == "query" {
			result, derr = model.QueryDecode(getStr("s"))
		} else {
			result, derr = model.PercentDecode(getStr("s"))
		}
		if derr != nil {
			return nil, derr
		}
		return text(fmt.Sprintf("decoded=%s", result)), nil

	case "query_list":
		qparams, err := model.ParseQuery(getStr("q"))
		if err != nil {
			return nil, err
		}
		var sb strings.Builder
		for _, p := range qparams {
			sb.WriteString(fmt.Sprintf("%s=%s\n", p.Key, p.Value))
		}
		return text(sb.String()), nil

	case "query_add":
		result := model.AddQueryParam(getStr("q"), getStr("key"), getStr("value"))
		return text(fmt.Sprintf("query=%s", result)), nil

	case "query_remove":
		result := model.RemoveQueryParam(getStr("q"), getStr("key"))
		return text(fmt.Sprintf("query=%s", result)), nil

	case "query_set":
		result := model.SetQueryParam(getStr("q"), getStr("key"), getStr("value"))
		return text(fmt.Sprintf("query=%s", result)), nil

	case "query_get":
		val, ok := model.GetQueryParam(getStr("q"), getStr("key"))
		if !ok {
			return nil, fmt.Errorf("key not found: %s", getStr("key"))
		}
		return text(fmt.Sprintf("key=%s value=%s", getStr("key"), val)), nil

	case "join":
		base := getStr("base")
		var paths []string
		if arr, ok := params.Arguments["paths"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					paths = append(paths, s)
				}
			}
		}
		result, err := model.JoinPath(base, paths...)
		if err != nil {
			return nil, err
		}
		return text(fmt.Sprintf("url=%s", result)), nil

	case "origin":
		p, err := model.Parse(getStr("url"))
		if err != nil {
			return nil, err
		}
		return text(fmt.Sprintf("origin=%s", p.Origin)), nil

	case "domain":
		p, err := model.Parse(getStr("url"))
		if err != nil {
			return nil, err
		}
		return text(fmt.Sprintf("domain=%s subdomain=%s tld=%s", p.Domain, p.Subdomain, p.TLD)), nil

	case "absolute":
		return text(fmt.Sprintf("absolute=%v", model.IsAbsolute(getStr("url")))), nil

	case "same_origin":
		result, err := model.SameOrigin(getStr("url1"), getStr("url2"))
		if err != nil {
			return nil, err
		}
		return text(fmt.Sprintf("same_origin=%v", result)), nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", params.Name)
	}
}
