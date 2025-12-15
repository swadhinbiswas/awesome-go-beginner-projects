# SMTP Verification Module

Purpose

Perform a safe, non-intrusive SMTP handshake to assess whether a mailbox
can receive mail. The module should never send message content and must be
rate-limited and auditable.

Preconditions

- Email syntax validated
- Domain validated (MX or A record exists)
- Domain not already classified as disposable (optional early exit)

1. SMTP Server Selection

- Use MX records ordered by priority. Attempt lower-numbered priorities
  first and failover to the next.
- If no MX, use RFC fallback (A record) only if allowed by policy.

2. TCP Connection Phase

- Connect to port 25 primarily; 587 as optional fallback only where
  permitted by policy.
- Apply strict connect and per-command timeouts.
- Failure handling:
  - connection refused → try next MX
  - timeout → Unknown
  - no MX reachable → mail server unreachable

3. SMTP Handshake (EHLO/HELO)

- Perform `EHLO` and parse server capabilities. Detect `STARTTLS` but do
  not require it for verification.
- Collect server banner and extension hints for telemetry.

4. MAIL FROM + RCPT TO

- Send a synthetic `MAIL FROM` (from a controlled domain) then `RCPT TO`
  for the target mailbox.
- Interpret responses:
  - `250` → mailbox exists (Valid)
  - `550` → mailbox does not exist (Invalid)
  - `252` → cannot verify (possible catch-all / privacy), treat as
    Unknown/Catch-All indication
  - `450/451` → temporary server failure (Unknown)

5. Catch-All Detection

- Strategy: also test one or two pseudo-random non-existent mailbox
  addresses and compare responses. If both the real and random addresses are
  accepted, flag as `Catch-All` (inconclusive for single-mailbox validity).

6. Anti-Verification Protections

- Many providers intentionally obscure mailbox existence (always return
  `252` or `250`). Recognize patterns:
  - always-accept behavior
  - greylisting / delayed responses
  - consistent `252` responses
- Treat such patterns as `Unknown` or `Risky` and avoid aggressive retries.

7. Timeouts & Retry Policy

- Use per-operation deadlines and an overall request timeout.
- Retry only by moving to the next MX host; never retry the same host in the
  same request.
- Limit total SMTP attempts per verification request.

8. Reputation & Abuse Protection

- Rate-limit SMTP checks per domain and per API key/IP.
- Cache SMTP failures and success states to reduce repeat probing.

9. Logging & Telemetry

- Log server banners, response codes, timeouts, and catch-all detection
  outcomes.
- Capture metrics: acceptance ratio, time-to-first-byte, per-domain failure
  spikes.

Outputs

- Mailbox status: `Valid` / `Invalid` / `Catch-All` / `Unknown`
- SMTP response code and server hostname
- Verification confidence delta (for scoring engine)

Security

- Never send message bodies.
- Do not attempt to harvest or verify more than policy allows.
- Respect server `STARTTLS` and other advertised policies; however, TLS may
  be skipped for quick probes depending on trust model.
