# Email Verifier – Go Mini Project

A modular **Email Verification Service** built as a **Golang mini project**. This project focuses on validating email addresses using multiple verification layers such as syntax checking, domain validation, DNS lookup, and SMTP probing—designed with clean architecture and real-world constraints in mind.

---

## 📌 Project Objective

The goal of this project is to determine whether an email address is:

* **Syntactically valid**
* **Associated with a real domain**
* **Capable of receiving emails**

The system produces a **verification status** and **confidence score** without sending any actual email content.

---

## 🧱 System Architecture (High-Level)

The service follows a **modular layered architecture**:

* API Layer
* Validation Engine
* DNS & SMTP Resolver Layer
* Scoring & Classification Layer
* Persistence & Caching Layer
* Security, Rate Limiting & Observability

Each module is independent, testable, and replaceable.

---

## 📂 Module Breakdown

### 1. Input Validation Module

**Purpose**

* Accept email input (single or batch)
* Normalize and sanitize input

**Responsibilities**

* Trim spaces and normalize case
* Validate email syntax (RFC-compliant)
* Reject malformed or empty inputs early

---

### 2. Domain Validation Module

**Purpose**

* Ensure the email domain exists and is reachable

**Responsibilities**

* Extract domain from email
* DNS lookup (A / AAAA records)
* MX record verification
* Detect disposable or temporary email domains

---

### 3. SMTP Verification Module

**Purpose**

* Verify mailbox existence without sending mail

**Responsibilities**

* Perform SMTP handshake
* Validate RCPT TO command
* Detect catch-all domains
* Apply strict timeouts and retries

**Safety Controls**

* No email body sent
* Rate-limited probing
* Blacklist-aware behavior

---

### 4. Scoring & Classification Module

**Purpose**

* Convert raw checks into meaningful results

**Statuses**

* Valid
* Invalid
* Risky
* Catch-All
* Unknown

**Outputs**

* Confidence score (percentage)
* Failure reason codes

---

### 5. Persistence & Caching Module

**Purpose**

* Improve performance and reduce redundant checks

**Stored Data**

* Email hash (never raw email)
* Domain name
* Verification result
* Timestamp
* Attempt count

---

### 6. API Layer

**Purpose**

* Expose verification functionality

**Conceptual Endpoints**

* Verify single email
* Verify batch emails
* Fetch verification history
* Health check

**Features**

* RESTful JSON API
* Versioned endpoints
* Consistent error responses

---

### 7. Rate Limiting & Security Module

**Purpose**

* Prevent abuse and protect SMTP reputation

**Controls**

* IP-based rate limiting
* API key authentication
* Request quotas
* Input sanitization

---

### 8. Logging & Observability Module

**Purpose**

* Monitor system health and failures

**Logs**

* DNS resolution errors
* SMTP failures
* Validation steps
* Rate-limit triggers

**Metrics**

* Success rate
* Average response time
* Domain-level failure patterns

---

## ⚙️ Concurrency & Performance

* Goroutine-based workers
* Controlled concurrency for SMTP checks
* Context-based cancellation
* Strict timeout enforcement per module


## 🧪 Testing Strategy

* Unit tests per module
---

**Decision Flow Diagram**

The decision flow diagram below visualizes the domain + SMTP verification
process. Open `decision-flow.svg` in this folder to view it.

![Decision Flow](decision-flow.svg)

**Packages (suggested)**

- `internal/extractor`: domain/local-part extraction and IDN→punycode normalization
- `internal/dnsresolver`: A/AAAA/MX lookups with timeouts, retries and TTLs
- `internal/disposable`: blacklist loader, wildcard/pattern matching and MX fingerprinting
- `internal/cache`: TTL-aware concurrency-safe cache (in-memory + optional Redis)
- `internal/classifier`: decision logic that merges DNS + disposable signals into final classification and confidence
- `internal/smtpclient`: safe SMTP probe wrapper (connect, EHLO, MAIL FROM, RCPT TO) with strict timeouts and anti-abuse controls
- `internal/policy`: runtime configuration for timeouts, retries, and probe toggles
- `internal/telemetry`: structured logging, metrics and tracing helpers
- `pkg/api` or `cmd/cli`: public integration surface (HTTP API or CLI)
- `internal/storage`: optional persistence adapters for verification history
- `scripts/test-stubs`: deterministic DNS & SMTP stubs for integration tests

**Recommended libraries & tools**

- `golang.org/x/net/idna` — IDN/punycode handling
- `miekg/dns` — advanced DNS control (optional)
- `net/textproto` / SMTP helpers or emersion libraries — SMTP interactions
- `patrickmn/go-cache` or `go-redis/redis` — caching backends
- `golang.org/x/time/rate` — rate limiting
- `prometheus/client_golang`, `uber-go/zap` — metrics & logging
- `mailhog` / `smtp4dev` — local SMTP stubs for testing

**What to learn / study**

- Go idioms: packages, interfaces, error handling, goroutines, channels and `context`
- Networking basics: TCP, DNS records (A, AAAA, MX), DNS TTL concepts
- SMTP protocol essentials: EHLO/MAIL FROM/RCPT TO and common response codes
- IDN and punycode normalization with `x/net/idna`
- Robustness patterns: timeouts, backoff strategies, retry policies
- Testing patterns: dependency injection, mocks, and deterministic stubs for DNS/SMTP
- Caching strategies: TTL-aware caching and invalidation
- Observability: structured logging, metrics (Prometheus), and tracing
- Security/ops: rate-limiting, avoiding abuse, and ethical SMTP probing

**What you'll learn by building this project**

- How to design clean, testable Go packages and interfaces
- Practical DNS and SMTP protocol knowledge and real-world edge cases
- Building resilient networked services using `context` and timeouts
- Creating deterministic integration tests with network stubs
- Implementing TTL-aware caching and efficient invalidation
- Instrumenting services for metrics, logging and alerting
- Operational concerns: rate-limits, IP reputation, and safe probing practices

---
# Email Verifier – Go Mini Project

This folder contains a design-first specification for an email verification
library written in Go. The intent is to keep the repository focused on a
clean, testable architecture and a clear implementation roadmap before any
production code is added.

Primary goals

- Determine domain existence and reachability
- Verify mail capability (MX / RFC fallback)
- Detect disposable / temporary domains (reputation-based)
- Optionally perform a non-intrusive SMTP handshake to probe mailbox existence

This repository provides the following detailed design documents:

- [`domain-validation.md`](./domain-validation.md) — Domain Validation Module: extraction, DNS checks,
	MX rules, disposable detection, caching and classification logic
- [`smtp-verification.md`](./smtp-verification.md) — SMTP Verification Module: safe handshake flow,
	timeout/retry rules, catch-all detection and anti-abuse strategies
- [`architecture.md`](./architecture.md) — Mapping into Go packages and interfaces, responsibilities
	per package, testing and dependency boundaries
- [`roadmap.md`](./roadmap.md) — Implementation roadmap, testing plan, CI suggestions and
	production hardening checklist

Usage

This folder is documentation-first. If you'd like, I can next:

- Generate Go package scaffolding and interface skeletons
- Produce a decision flow diagram (SVG/PNG) representing the verification flow
- Create a minimal runnable prototype (CLI) that exercises the Domain module

Tell me which of the above you'd like to continue with and I will scaffold it.
* ML-based reputation scoring
