package mux

import (
	"net/http"
	"webserver/routes"
)

// NewRouter creates and configures a new http.ServeMux
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", routes.IndexHandler)
	mux.HandleFunc("/about", routes.AboutHandler)
	mux.HandleFunc("/blog", routes.BlogHandler)
	mux.HandleFunc("/docs", routes.DocsHandler)
	mux.HandleFunc("/gallery", routes.GalleryHandler)
	mux.HandleFunc("/services", routes.ServicesHandler)
	mux.HandleFunc("/signup", routes.SignupHandler)
	mux.HandleFunc("/login", routes.LoginHandler)
	mux.HandleFunc("/dashboard", routes.AuthMiddleware(routes.DashboardHandler))
	mux.HandleFunc("/logout", routes.LogoutHandler)



	// Serve other static assets (css, js, images) if they were in a subdirectory,
	// but here the html files are in the root of static/.
	// If the html files reference other assets, we might need a file server.
	// For now, this covers the pages.

	return mux
}
