# Master Project Roadmap: Email Verification Service

This document serves as the comprehensive guide to building the **Email Verification Service**. It synthesizes the architectural decisions, validation logic, and safety protocols defined in the project's design documents (`architecture.md`, `domain-validation.md`, `smtp-verification.md`).

**Constraint:** No code is used in this guide. Focus is entirely on logic, flow, and system design.

*(See also: [`architecture.md`](./architecture.md), [`domain-validation.md`](./domain-validation.md), [`smtp-verification.md`](./smtp-verification.md))*

---

## 1. Project Philosophy & Architecture
**Why we built it like this:**
We avoided a monolithic script in favor of a **Clean Architecture**. This allows us to separate the "Business Logic" (Is this email valid?) from the "Infrastructure" (DNS queries, SMTP connections).

### The Layered Design
We will build the system in distinct layers, as defined in `architecture.md`:

1.  **Input Layer:** Handles raw string inputs (API/CLI).
2.  **Extractor & Normalizer:** Cleans input and handles international domains (Punycode).
3.  **Validation Engine (The Core):** Orchestrates the checks.
4.  **Resolver Layer (Infrastructure):**
    *   **DNS Resolver:** Handles looking up A, AAAA, and MX records.
    *   **Disposable Detector:** Checks allowlists/blocklists.
    *   **SMTP Client:** Speaks the raw mail protocol.
5.  **Persistence Layer:** Caches results to prevent redundant network calls.

---

## 2. Prerequisites
Before beginning, you must understand the following core concepts:

*   **DNS Records:**
    *   **A / AAAA:** Maps a hostname to an IP address (IPv4/IPv6).
    *   **MX (Mail Exchange):** Specifies the mail servers responsible for accepting email on behalf of a domain.
    *   **TTL (Time To Live):** How long a DNS record should be cached.
*   **SMTP Protocol:**
    *   The conversational flow: `EHLO` -> `MAIL FROM` -> `RCPT TO` -> `QUIT`.
    *   Response codes: `250` (OK), `550` (User Unknown), `4xx` (Temporary Failure).
*   **Network Resilience:**
    *   **Timeouts:** Implementing strict deadlines for every network operation.
    *   **Backoff:** Waiting increasingly longer intervals before retrying failures.

---

## 3. Implementation Phases

### Phase 1: Input & Domain Logic (The Foundation)
*Reference: [`domain-validation.md`](./domain-validation.md)*

**Objective:** Sanitize input and verify if the domain acts like a real internet citizen.

**Step 1.1: Extraction & Normalization**
*   **Logic:** Split the email at the `@` symbol.
*   **Normalization:** Convert the domain to lowercase. Crucially, process **Internationalized Domain Names (IDN)**. If a user enters `ñ`, it must be converted to its ASCII "Punycode" equivalent (e.g., `xn--...`) so DNS servers understand it.

**Step 1.2: DNS Lookups**
We perform three queries in parallel or sequence:
1.  **MX Lookup:** The primary signal. Does this domain handle mail?
2.  **A/AAAA Fallback:** If NO MX records exist, does the domain translate to an IP? (RFC standard allows sending mail to the A record if MX is missing).
3.  **Validation Logic:**
    *   **Valid:** MX records found.
    *   **Fallback Valid:** No MX, but A record found.
    *   **Invalid:** NXDOMAIN (domain doesn't exist) or no records at all.
    *   **Null MX:** If the MX record is explicitly `.`, the domain is signaling "We do not accept mail." Mark as Invalid immediately.

**Step 1.3: Disposable Detection**
*   **Static Lists:** Check against a known list of "10-minute mail" providers.
*   **Pattern Matching:** Detect wildcard domains used by spammers.
*   **MX Fingerprinting:** If a domain uses a mail server known to host disposable emails (e.g., specific AWS or DigitalOcean endpoints often reused for spam), flag it as **Risky**.

---

### Phase 2: SMTP Verification (The Network Probe)
*Reference: [`smtp-verification.md`](./smtp-verification.md)*

**Objective:** Connect to the mail server to verify the specific user, safely.

**Step 2.1: Server Selection**
*   Sort the MX records by priority (lowest number first).
*   Attempt to connect to the top priority server. If it fails (timeout/refused), failover to the next one.

**Step 2.2: The Handshake**
1.  **Connect:** Open a TCP connection strictly on Port 25 (or 587 if configured).
2.  **EHLO:** Send the standard greeting.
3.  **MAIL FROM:** We must say who *we* are. Use a neutral sender address (e.g., `verifier@<our-hostname>`).
4.  **RCPT TO:** The critical question. We present the target email address.

**Step 2.3: Interpreting the Response**
*   **250 OK:** The user likely exists.
*   **550 User Unknown:** The user definitely does not exist.
*   **252:** The server accepted it but won't confirm delivery (Privacy/Security feature). Treat as "Unknown".
*   **4xx:** Temporary error (e.g., Greylisting). Do not mark as Invalid; mark as Unknown.

**Step 2.4: Catch-All Detection**
*   **The Trap:** Some servers say "250 OK" for *every* email to prevent spam harvesting.
*   **The Counter-Move:** If `RCPT TO` succeeds, we perform a "Catch-All Probe". We immediately send a second `RCPT TO` with a random, non-existent user (e.g., `fsdjkfhuew@domain.com`).
*   **Logic:**
    *   If the random user is **Rejected** (550) but our target is **Accepted** (250) -> The target is **Valid**.
    *   If *both* are Accepted (250) -> The domain is a **Catch-All**. The specific email cannot be verified solely via SMTP.

---

### Phase 3: Classification & Caching (The Decision)
*Reference: [`architecture.md`](./architecture.md)*

**Objective:** Synthesize all signals into a final result and save it.

**Step 3.1: Scoring Logic**
The system outputs a detailed status, not just a boolean:
*   **Deliverable:** DNS Valid + SMTP confirms user + Not Catch-All.
*   **Risky:** Catch-All detected OR Disposable pattern matched.
*   **Undeliverable:** DNS NXDOMAIN OR SMTP 550 User Unknown.
*   **Unknown:** SMTP Timeout or Greylisting.

**Step 3.2: Caching Strategy**
*   **Keys:** Cache by the normalized domain/email.
*   **TTL:** Respect the DNS Time-To-Live. Do not cache "Unknown" (timeouts) for long, as they might be transient.
*   **Why:** Prevents us from hammering the same mail server repeatedly (Abuse prevention).

---

## 4. Learning Outcomes
By building this systems-level project, you will master:

1.  **RFC Compliance:** Understanding that real-world protocols (SMTP) are messy and require handling many edge cases (Greylisting, Tarpitting).
2.  **Defensive Networking:** How to write code that assumes the network *will* fail (handling timeouts, connection resets).
3.  **Punycode & Encodings:** Managing international text standards.
4.  **System Design:** Designing stateless modules (Validator, Resolver) that can be tested in isolation using **Dependency Injection**.
5.  **Ethical Scraping:** Learning how to verify data without being abusive (Rate limiting, respecting server policies).
