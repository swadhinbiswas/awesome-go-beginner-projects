Fix: Protect routes so they cannot be accessed without login

Goal

- Ensure particular routes (for example `/dashboard`) are accessible only after a successful login.
- Provide a minimal, clear implementation you can drop into the existing project.

Overview of approaches (short)

- Simple signed cookie: store a signed user ID in a cookie using `securecookie` or `gorilla/sessions`.
- Server-side session store: store session id in cookie and keep session data (user id) in Redis/DB. More secure and scalable.
- JWT: return a signed JWT token to the client and validate it for protected routes. Good for APIs, more complexity.

Recommended minimal implementation (development-friendly)

- Use a signed cookie to store the authenticated user's ID. This is easy to add and works without an external store.
- Steps below include code snippets that integrate with current files in the repo:
  - `routes/routes.go` (login handler and middleware)
  - `mux/mux.go` (wrap protected routes)
  - new `routes/logout` handler to clear the cookie

1. Add securecookie helper (install dependency)

For a minimal signed cookie, use `github.com/gorilla/securecookie` which is simple and widely used.

Add to your module (run in project root):

```bash
cd /home/swadhin/JOB/LEARNGO/awesome-go-beginner-projects/webserver
go get github.com/gorilla/securecookie
```

2. Setup a package-level securecookie instance

In `functionality` (or `routes`) create a small helper. Example (put in `functionality/auth.go`):

```go
package functionality

import (
    "github.com/gorilla/securecookie"
)

// Use strong keys in production. Here they are generated for demo.
var (
    // hashKey authenticates values
    hashKey = []byte("very-secret-hash-key-32-bytes-minimum")
    // blockKey encrypts values (optional). Use nil if not encrypting.
    blockKey = []byte(nil)
    S = securecookie.New(hashKey, blockKey)
)
```

3. Set the cookie on successful login

In `routes/LoginHandler` (after verifying password) set a signed cookie with the user id:

```go
import (
    "net/http"
    "time"
    "webserver/functionality"
)

// inside LoginHandler, after VerifyUserPassword succeeds:
value := map[string]string{
    "user_id": user.ID,
}
encoded, err := functionality.S.Encode("session", value)
if err != nil {
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    return
}
http.SetCookie(w, &http.Cookie{
    Name:     "session",
    Value:    encoded,
    Path:     "/",
    HttpOnly: true,
    Secure:   false, // true in production over HTTPS
    SameSite: http.SameSiteLaxMode,
    Expires:  time.Now().Add(24 * time.Hour),
})
// then redirect to /dashboard
http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
```

Notes:

- `Secure: true` should be set when your site runs over HTTPS.
- `HttpOnly` prevents JS access to the cookie.
- `SameSite` reduces CSRF exposure for cross-site requests.

4. Implement middleware to verify the cookie

You already have an `authMiddleware` function skeleton in `routes/routes.go`. Replace its simple cookie check with securecookie decoding and optional DB verification:

```go
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session")
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        value := make(map[string]string)
        if err := functionality.S.Decode("session", cookie.Value, &value); err != nil {
            // Invalid or tampered cookie
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        userID := value["user_id"]
        if userID == "" {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Optional: load user from DB to ensure the user still exists
        var user functionality.User
        if err := functionality.DB.First(&user, "id = ?", userID).Error; err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Optionally set the user in request context for handlers to use
        // ctx := context.WithValue(r.Context(), "user", &user)
        // next(w, r.WithContext(ctx))

        next(w, r)
    }
}
```

5. Wrap protected handlers with the middleware in `mux/mux.go`

Instead of registering `routes.DashboardHandler` directly, wrap it:

```go
mux.HandleFunc("/dashboard", routes.AuthMiddleware(routes.DashboardHandler))
```

Note: the existing `authMiddleware` in your repo uses lowercase name and signature `authMiddleware(next http.HandlerFunc) http.HandlerFunc`. You can export it as `AuthMiddleware` if you prefer using it from other packages or just call `routes.authMiddleware` from `mux` when `routes` is in same package (it's the same package already). Keep naming consistent.

6. Add a `/logout` endpoint to clear the cookie

```go
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    // Clear the cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "session",
        Value:    "",
        Path:     "/",
        Expires:  time.Unix(0, 0),
        MaxAge:   -1,
        HttpOnly: true,
        Secure:   false,
    })
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}
```

Register in `mux/mux.go`:

```go
mux.HandleFunc("/logout", routes.LogoutHandler)
```

7. Optionally: store session server-side (recommended for production)

- Instead of embedding the user id (even signed) in the cookie, store a random session ID in the cookie and keep session data (user id, expiry) in Redis or database. This allows server-side session revocation.
- For that approach: generate session token, save it with associated user id and expiry in Redis/DB, set the cookie to the token value.

8. Security considerations

- Always use strong random keys for `securecookie.New`:
  - `hashKey` must be at least 32 bytes. Use `github.com/gorilla/securecookie.GenerateRandomKey(32)` to generate and store safely (not in source control).
- Use `Secure: true` in cookies when running on HTTPS in production.
- Use `HttpOnly` to avoid access from JS.
- Set `SameSite=Lax` or `Strict` depending on needs.
- Consider server-side sessions for revocation, logout safety, and short-lived tokens.
- Consider CSRF protection for state-changing POST endpoints (e.g., forms), either via SameSite cookies + CSRF token, or using double-submit pattern.

9. Testing steps

- Start the server:

```bash
go run .
```

- Create a user via `/signup`.
- Login and inspect cookies in browser DevTools -> Application -> Cookies -> your host. You should see the `session` cookie.
- Visit `/dashboard` in a private/incognito window (or after clearing cookies) — you should be redirected to `/login`.
- Click Logout to ensure the cookie is removed and you can no longer reach `/dashboard` without login.

10. Example quick code diffs (summary)

- `functionality/auth.go` (new) - add `securecookie` instance
- `routes/routes.go` - set cookie in `LoginHandler`, strengthen `authMiddleware`, add `LogoutHandler`
- `mux/mux.go` - register `/logout` and wrap protected routes with middleware

If you want, I can apply these exact code changes to the repository now (I will:

- add `functionality/auth.go` with securecookie config,
- update `routes/LoginHandler` to set the signed cookie,
- update `authMiddleware` to use `securecookie` decoding,
- add `LogoutHandler`, and
- register `/logout` in `mux/mux.go`).

Which option do you want next?

- "Apply changes now" — I'll patch the files and run quick compile checks.
- "Only document" — keep this file only and you will apply the changes yourself.
- "Use server-side sessions" — I'll implement a simple in-memory session store (not for production) to demonstrate.
