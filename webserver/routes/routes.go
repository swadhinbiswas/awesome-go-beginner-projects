package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"webserver/functionality"
)

func IndexHandler(w http.ResponseWriter,
	r *http.Request) {
	if r.URL.Path != "/" {

		ErrorHandler(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("static", "index.html"))
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "about.html"))
}

func BlogHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "blog.html"))
}

func DocsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "docs.html"))
}

func GalleryHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "gallery.html"))
}

func ServicesHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "services.html"))
}

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(filepath.Join("static", "404.html"))
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write(content)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, filepath.Join("static", "signup.html"))
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if email == "" || username == "" || password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if password != confirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	newUser := functionality.User{
		Email:    email,
		Username: username,
		Password: password,
	}
	result := functionality.DB.Create(&newUser)
	if result.Error != nil {
		http.Error(w, "Error creating user: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("User created successfully")
	http.ServeFile(w, r, filepath.Join("static", "signup_success.html"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, filepath.Join("static", "login.html"))
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Support both application/x-www-form-urlencoded (regular HTML forms)
	// and application/json (API clients). Detect content-type and parse accordingly.
	ct := r.Header.Get("Content-Type")
	var identifier, password string

	if strings.Contains(ct, "application/json") {
		var payload struct {
			Identifier string `json:"identifier"`
			Password   string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Error parsing JSON", http.StatusBadRequest)
			return
		}
		identifier = strings.TrimSpace(payload.Identifier)
		password = payload.Password
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		identifier = strings.TrimSpace(r.FormValue("identifier"))
		password = r.FormValue("password")
	}

	// Helpful debug logging while diagnosing client issues
	log.Printf("Login attempt Content-Type=%q identifier=%q", ct, identifier)

	if identifier == "" || password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	var user functionality.User
	if err := functionality.DB.
		Where("email = ? OR username = ?", identifier, identifier).
		First(&user).Error; err != nil {

		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := functionality.VerifyUserPassword(&user, password); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Create a signed session token (expiry 24h) and set cookie.
	token, err := functionality.CreateSessionToken(user.ID, 24*time.Hour)
	if err != nil {
		log.Printf("failed to create session token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set to true in production (HTTPS)
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// On successful login redirect to dashboard (GET)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// LogoutHandler clears the session cookie and redirects to login.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
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

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userID, ok := functionality.ParseSessionToken(cookie.Value)
		if !ok || userID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Optional: verify user exists in DB
		var u functionality.User
		if err := functionality.DB.First(&u, "id = ?", userID).Error; err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}


// DashboardHandler serves a simple dashboard page after successful login.
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, filepath.Join("static", "dashboard.html"))
}


func ProfileHandler(w http.ResponseWriter, r *http.Request) {
// To be implemented: serve user profile page
userID, ok := getUserIDFromRequest(r)
if !ok {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return
}
var profile functionality.Profile
if err := functionality.DB.
		Where("user_id = ?", userID).
		First(&profile).Error; err != nil {
	http.Error(w, "Profile not found", http.StatusNotFound)
	return

}
http.ServeFile(w, r, filepath.Join("static", "profile.html"))
}

func getUserIDFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		return "", false
	}

	userID, ok := functionality.ParseSessionToken(cookie.Value)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}

func EditProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		http.ServeFile(w, r, filepath.Join("static", "edit_profile.html"))
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	bio := r.FormValue("bio")||""
	avaterUrl := r.FormValue("avater_url")||""
	name:= r.FormValue("name")||""

	var profile functionality.Profile
	if err := functionality.DB.
		Where("user_id = ?", userID).
		First(&profile).Error; err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	profile.Bio = bio
	profile.AvaterUrl = avaterUrl
	profile.Name = name


	if err := functionality.DB.Save(&profile).Error; err != nil {
		http.Error(w, "Error updating profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
