# Domain Validation Module

Purpose

The Domain Validation Module determines whether the domain portion of an
email address is real, reachable on the network, capable of receiving mail,
and not part of a disposable/temporary provider. This module runs before any
SMTP probing and produces deterministic signals used by higher-level
classification and scoring.

1. Domain Extraction

- How: Split the input on the single `@` character and validate there is
  exactly one `@`.
- Normalize: convert to lowercase, strip trailing dots, and normalize
  Internationalized Domain Names (IDN) to punycode representation.
- Why: prevents malformed DNS queries and ensures cache keys are consistent.

2. DNS Lookup (A / AAAA Records)

- Purpose: verify the domain resolves to an address (network-reachable).
- Operations: perform A (IPv4) and AAAA (IPv6) lookups using Go's standard
  resolver.
- Decision logic:
  - if any A or AAAA record exists → consider domain network-reachable
  - NXDOMAIN → domain does not exist (Invalid)
  - Timeout → mark result Unknown
  - SERVFAIL → retry once with exponential (or linear) backoff before
    escalating

3. MX Record Verification

- Purpose: ensure the domain advertises mail exchangers (MX records) that
  can receive mail.
- How:
  - Query MX records and sort by preference
  - Validate that each MX target resolves (A/AAAA) or is otherwise
    routable
- Decision logic:
  - MX exists and valid targets → Mail-capable
  - No MX but A exists → RFC fallback allowed (treat as mail-capable)
  - No MX and no A → Invalid domain
  - Null MX (a record of `.`) → explicit refusal to accept mail (Invalid)

4. Disposable / Temporary Email Detection

- Core approach: reputation- and pattern-based detection (not pure DNS).
- Techniques:
  - Static domain blacklist (JSON/text/embedded resource)
  - Wildcard and pattern matching to handle rotating subdomains
  - MX provider fingerprinting — detect disposable providers by MX
    hostname patterns
  - Heuristics (advanced): extremely low DNS TTLs, frequently changing MX
    entries, short registration age — mark as Risky, not definitive

- Flow:
  - Extract normalized domain
  - Check static blacklist → immediate Disposable if matched
  - Run pattern and MX-fingerprint checks → Disposable or Risky

5. Result Classification

- No DNS records → Invalid
- A exists, MX exists → Valid
- A exists, no MX → Valid (RFC fallback)
- Null MX → Invalid
- Disposable domain match → Disposable
- DNS timeout → Unknown
- Suspicious heuristics → Risky

6. Caching Strategy

- Use caching for DNS results and disposable checks to avoid repeated
  lookups. Cache keys are the normalized punycode domain.
- Respect DNS TTL for cache invalidation. For blacklist/pattern data, use
  a configurable refresh interval.

7. Failure Handling & Best Practices

- Always return a classification (never hang forever).
- Use `context` timeouts per operation.
- Retry DNS once on transient errors.
- Do not brute-force DNS resolution; rate-limit per domain and per client.

8. Security & Abuse Considerations

- Rate-limit checks, especially DNS-heavy operations.
- Avoid repeatedly querying known-malicious domains.
- Log and surface metrics for high-failure domains so operators can tune
  blacklists.

Outputs

- Domain validity: `Valid` / `Invalid` / `Unknown`
- Mail capability: `Yes` / `No` (or `Fallback`)
- Disposable flag: `True` / `False` / `Risky`
- Confidence impact score: numeric delta used by the scorer

Notes on implementing in Go

- Use Go's `net` package resolver or a configurable resolver (system vs
  custom DNS server) depending on deployment needs.
- Wrap lookups with contexts and strict timeouts.
- Keep disposable lists as configurable resources and support hot reloads.
