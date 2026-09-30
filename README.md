# urlkit

Agentic-first URL parsing and manipulation service. Parse URLs into components, build URLs from parts, manipulate query parameters, normalize URLs, resolve relative URLs, percent-encode/decode, extract domains. Plain text API, agent-driven, single Go binary.

## Quick Start

```bash
make build
./urlkit
```

Then:

```bash
# Parse a URL
curl "http://localhost:8080/parse?url=https://sub.example.com:8080/api/v1?q=1#top"

# Build a URL
curl "http://localhost:8080/build?scheme=https&host=example.com&path=/api&query=a=1"

# Resolve a relative URL
curl "http://localhost:8080/resolve?base=https://example.com/a/b/c&ref=../d"

# Normalize a URL
curl "http://localhost:8080/normalize?url=https://Example.com:443/a//b/?z=2&a=1#frag"

# Get JSON
curl -H "Accept: application/json" "http://localhost:8080/parse?url=https://example.com"
```

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/parse?url=` | Parse URL into components |
| GET | `/build?scheme=&host=&path=&query=&fragment=` | Build URL from parts |
| GET | `/resolve?base=&ref=` | Resolve relative URL |
| GET | `/normalize?url=` | Normalize URL |
| GET | `/encode?s=&mode=path\|query` | Percent-encode string |
| GET | `/decode?s=&mode=path\|query` | Percent-decode string |
| GET | `/query?action=list\|add\|remove\|set\|get&q=&key=&value=` | Query param manipulation |
| GET | `/join?base=&path=&path=` | Join path segments |
| GET | `/origin?url=` | Extract origin |
| GET | `/domain?url=` | Extract domain/subdomain/TLD |
| GET | `/absolute?url=` | Check if absolute |
| GET | `/same-origin?url1=&url2=` | Check same origin |
| POST | `/mcp` | MCP JSON-RPC 2.0 endpoint |
| GET | `/help` | Operating manual |

## Configuration

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `-addr` | `URLKIT_ADDR` | `:8080` | Listen address |

## Principles

- **Agent IS the interface** — No UI, no SDK
- **Plain text by default** — JSON on demand via `Accept: application/json` or `?format=json`
- **Instructive errors** — Every 4xx includes a hint
- **Self-documenting** — `GET /help` returns the operating manual
- **Single static binary** — CGO_ENABLED=0, zero external dependencies
- **No database** — Pure stateless computation

## License

MIT
