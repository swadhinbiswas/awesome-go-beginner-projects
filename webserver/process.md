Process: Protect routes using only Go standard library (no third-party)

Summary

- Goal: ensure particular routes (e.g. `/dashboard`) are accessible only after a successful login.
- Approach: create an HMAC-signed session token (userID + expiry), store it in an HttpOnly cookie, validate in middleware.
- No third-party packages required; only Go stdlib (`crypto/hmac`, `crypto/sha256`, `crypto/subtle`, `encoding/base64`, `crypto/rand`, `net/http`, etc.).

Step-by-step process

1. Choose token format

- Format: `payload.signature` where:
  - `payload` = base64("userID:expiryUnix")
  - `signature` = base64(HMAC-SHA256(secret, payloadRaw))
- Payload must contain an expiry timestamp to prevent indefinite reuse.

2. Generate and persist a secret key

- Generate a strong random key (>= 32 bytes). Persist it as an environment variable (e.g. `SESSION_KEY_BASE64`).
- Example (generate and print base64):

```bash
head -c 32 /dev/urandom | base64
```

- In your app startup, read `SESSION_KEY_BASE64`, decode base64 to bytes, and use that as the HMAC secret. If missing, log a warning and generate an ephemeral key (sessions will not survive restarts).

3. Create a session token after successful login

- After verifying username/email + password:
  - Build payload: `userID:expiryUnix` (expiry = now + TTL, e.g. 24h).
  - Compute HMAC-SHA256 over the raw payload (not the base64) using secret key.
  - Encode token: `base64(payloadRaw) + "." + base64(signature)`.
  - Set cookie on response with that token.

4. Set secure cookie attributes

- Use `http.SetCookie` with:
  - `Name`: `session`
  - `Value`: token
  - `Path`: `/`
  - `HttpOnly`: true
  - `Secure`: true in production (HTTPS)
  - `SameSite`: `http.SameSiteLaxMode` or stricter
  - `Expires` / `MaxAge`: match TTL

5. Validate token in middleware (protect routes)

- Implement middleware that runs before protected handlers:
  - Read `session` cookie.
  - Split token on `.`; base64-decode payload and signature.
  - Recompute expected HMAC from payload and secret; compare signatures using `crypto/subtle.ConstantTimeCompare`.
  - Parse payload to extract `userID` and `expiryUnix`; verify `expiryUnix` > now.
  - Optionally load user from DB to ensure it still exists.
  - If everything checks out, call `next(w, r)`, otherwise redirect to `/login` (or return 401 for API clients).

6. Provide logout

- Implement `LogoutHandler` that sets an expired cookie (Value empty, Expires in the past, MaxAge -1) and redirects to `/login`.

7. Optional: server-side session store (recommended for revocation)

- For immediate revocation (logout everywhere) use a server-side session store:
  - Generate random session ID, persist it with userID + expiry in Redis/DB.
  - Put session ID in the cookie (signed or plain if store is authoritative).
  - Middleware checks session ID against store and enforces expiry and revocation.
- In-memory store OK for demos only; use Redis or DB for production.

8. Security checklist and best practices

- Never hard-code secret keys in source control; use env vars/secret manager.
- Use at least 32 random bytes for the HMAC key.
- Always compare MACs with constant-time compare to avoid timing attacks.
- Use `Secure: true` for cookies in production HTTPS deployments.
- Use `HttpOnly` to prevent access from JavaScript.
- Rotate keys safely: support verifying tokens with older keys for a migration window.
- Consider encrypting payload with AES-GCM if you must hide userID or other contents.
- Protect state-changing POST endpoints against CSRF (SameSite helps, but consider CSRF tokens for full protection).

9. Commands & quick checks

- Generate key:

```bash
head -c 32 /dev/urandom | base64 > session_key.b64
export SESSION_KEY_BASE64=$(cat session_key.b64)
```

- Test login flow from shell (form-encoded):

```bash
curl -v -X POST -d "identifier=alice&password=secret" http://localhost:8080/login
```

- Test JSON login:

```bash
curl -v -X POST -H "Content-Type: application/json" -d '{"identifier":"alice","password":"secret"}' http://localhost:8080/login
```

- After login, inspect cookie in browser DevTools or from curl using `-v` to see `Set-Cookie` header.

10. Where to implement in this repo (file map)

- `functionality/` (new file, e.g. `auth_std.go`): helpers to sign/verify and to read the env var key.
- `routes/routes.go`: after successful login create token and `http.SetCookie`; add `LogoutHandler`; implement `AuthMiddleware` which validates token.
- `mux/mux.go`: register protected routes wrapped with middleware, e.g. `mux.HandleFunc("/dashboard", routes.AuthMiddleware(routes.DashboardHandler))`; register `/logout`.

11. Testing plan

- Create a user with `/signup`.
- Login and confirm `Set-Cookie` appears and cookie has correct attributes.
- Access `/dashboard` with cookie — should work.
- Access `/dashboard` after clearing cookies — should redirect to `/login`.
- Manually tamper cookie or change signature — should be rejected.
- If using server-side sessions, delete session server-side and confirm access denied.

Notes

- This approach uses only Go standard library primitives; it is secure if implemented correctly and keys are managed safely.
- For production systems that require revocation, scaling, or advanced features, consider a server-side store or a vetted library; but the stdlib approach is perfectly acceptable for many smaller apps.

If you want, I can now either:

- produce small pseudo-code snippets for each step (still no full file edits), or
- apply the minimal stdlib implementation to the repository files for you.
