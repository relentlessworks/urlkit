package api

const helpText = `urlkit — Agentic-First URL Parsing & Manipulation Service

No auth required. All endpoints are stateless. Plain text by default.
JSON via Accept: application/json header or ?format=json query param.

ENDPOINTS

  GET /parse?url=<URL>
    Decompose a URL into scheme, host, port, path, query, fragment, domain, etc.
    Example: GET /parse?url=https://sub.example.com:8080/api/v1?q=1#top
    Output:  scheme=https host=sub.example.com:8080 hostname=sub.example.com port=8080 path=/api/v1 query=q=1 fragment=top
             domain=example.com subdomain=sub tld=com origin=https://sub.example.com:8080
             is_https=true is_secure=true has_port=true has_query=true has_fragment=true

  GET /build?scheme=<s>&host=<h>&path=<p>&query=<q>&fragment=<f>&username=<u>&password=<pw>
    Construct a URL from individual components.
    Example: GET /build?scheme=https&host=example.com&path=/api&query=a=1
    Output:  url=https://example.com/api?a=1

  GET /resolve?base=<URL>&ref=<ref>
    Resolve a relative URL reference against a base URL (like a browser).
    Example: GET /resolve?base=https://example.com/a/b/c&ref=../d
    Output:  url=https://example.com/a/d

  GET /normalize?url=<URL>
    Normalize: lowercase scheme/host, remove default ports, dedup slashes,
    remove trailing slash, sort query params, remove fragment.
    Example: GET /normalize?url=https://Example.com:443/a//b/?z=2&a=1#frag
    Output:  url=https://example.com/a/b?a=1&z=2

  GET /encode?s=<string>&mode=<path|query>
    Percent-encode a string for use in a URL path or query value.
    Example: GET /encode?s=hello world&mode=path
    Output:  encoded=hello%20world

  GET /decode?s=<encoded>&mode=<path|query>
    Percent-decode a string.
    Example: GET /decode?s=hello%20world
    Output:  decoded=hello world

  GET /query?action=<list|add|remove|set|get>&q=<query>&key=<k>&value=<v>
    Manipulate query parameters.
    list:   GET /query?action=list&q=a=1&b=2
    add:    GET /query?action=add&q=a=1&key=b&value=2
    remove: GET /query?action=remove&q=a=1&b=2&key=a
    set:    GET /query?action=set&q=a=1&b=2&key=a&value=3
    get:    GET /query?action=get&q=a=1&b=2&key=a

  GET /join?base=<URL>&path=<seg1>&path=<seg2>
    Join path segments to a base URL.
    Example: GET /join?base=https://example.com/api&path=v1&path=users
    Output:  url=https://example.com/api/v1/users

  GET /origin?url=<URL>
    Extract the origin (scheme://host) from a URL.
    Example: GET /origin?url=https://example.com/path
    Output:  origin=https://example.com

  GET /domain?url=<URL>
    Extract domain, subdomain, and TLD from a URL.
    Example: GET /domain?url=https://sub.example.com/path
    Output:  domain=example.com subdomain=sub tld=com

  GET /absolute?url=<URL>
    Check if a URL is absolute (has a scheme).
    Example: GET /absolute?url=https://example.com
    Output:  absolute=true

  GET /same-origin?url1=<URL>&url2=<URL>
    Check if two URLs share the same origin (scheme + host).
    Example: GET /same-origin?url1=https://a.com&url2=https://a.com/path
    Output:  same_origin=true

  POST /mcp
    MCP JSON-RPC 2.0 endpoint. Tools: parse, build, resolve, normalize,
    encode, decode, query_list, query_add, query_remove, query_set, query_get,
    join, origin, domain, absolute, same_origin.

  GET /help (or /.well-known/agent.md)
    This help text.

ERRORS
  All 4xx responses: error: <message> | hint: <what to do next>
`
