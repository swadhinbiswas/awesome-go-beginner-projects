# Architecture — Mapping to Go Packages & Interfaces

This document maps the domain and SMTP module responsibilities into clean
Go packages and describes the key interfaces, responsibilities and testing
boundaries. The guidance below is deliberately language-idiomatic but avoids
implementation code: it focuses on responsibilities and contracts.

Package layout (recommended)

- `internal/extractor` — Domain and local-part extraction & normalization
- `internal/dnsresolver` — DNS lookups (A, AAAA, MX) with timeout + retry
- `internal/disposable` — Disposable provider detection (blacklist,
  pattern-matching, MX fingerprinting)
- `internal/classifier` — Classification logic that combines DNS & disposable
  signals into final domain status
- `internal/cache` — TTL-aware cache for DNS & disposable results
- `internal/smtpclient` — Safe SMTP probe wrapper with timeouts, retries,
  and anti-abuse controls
- `cmd/cli` or `pkg/api` — Integration surface (CLI or HTTP API)
- `internal/telemetry` — Logging, metrics and tracing helpers

Key interface responsibilities (described)

- Domain extractor
  - Responsibility: take raw input and produce a normalized domain token
    plus error cases for malformed inputs. Must guarantee idempotent output
    for equal inputs.

- DNS resolver
  - Responsibility: perform A/AAAA/MX lookups with contexts and per-call
    timeouts. Must return record sets plus effective TTLs for cache use and
    classify transient errors (timeout / servfail) vs permanent (nxdomain).

- Disposable checker
  - Responsibility: check whether a domain (or its MX) is a known disposable
    provider. Uses local resources (blacklist) and heuristics. Should support
    pattern matching and runtime updates.

- Cache
  - Responsibility: store and retrieve DNS/disposable results indexed by
    normalized domain. Must respect TTLs and provide safe concurrency.

- Classifier
  - Responsibility: combine resolver + disposable signals and produce a
    domain classification (Valid/Invalid/Unknown/Disposable/Risky) and a
    confidence delta. Stateless business logic that depends on inputs only.

- SMTP client
  - Responsibility: connect, EHLO, MAIL FROM, RCPT TO and interpret replies
    safely. Must accept context with overall deadlines and expose events for
    telemetry. Rate-limiting and retry behavior should be enforced externally
    by the caller or via an injected policy object.

Design & integration guidelines

- Use dependency injection: pass interfaces into the classifier/engine so
  tests can inject deterministic behaviors (e.g., mock resolvers and
  smtpclients).
- Keep side-effecting operations (network, logging) in thin packages so the
  classifier logic remains pure and easy to unit test.
- Make caching pluggable. For local development an in-memory cache is fine;
  production can use Redis or similar with TTL-awareness.
- Use contexts broadly and enforce timeouts at package boundaries.

Testing strategy per package

- `extractor`: unit tests for normalization and IDN handling
- `dnsresolver`: table-driven tests plus integration tests against a
  deterministic DNS stub; exercise NXDOMAIN, SERVFAIL, TTLs
- `disposable`: unit tests for blacklist lookup, wildcard pattern matches,
  and MX-fingerprint heuristics
- `classifier`: pure unit tests combining stubbed inputs to validate
  classification matrix
- `smtpclient`: integration tests against a local SMTP stub/emulator that
  returns scripted responses for `RCPT TO` and `EHLO`

Observability & safety

- Emit structured logs for transient failures and timeouts.
- Expose metrics for DNS resolution times, SMTP successes/failures, and
  disposable hits.
- Add guardrails: global rate limits, per-domain backoffs, and cache-based
  cooldowns for repeated failures.
