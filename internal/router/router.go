package router

import (
	"net/http"

	"github.com/danqzq/mdspace/internal/handlers"
	"github.com/danqzq/mdspace/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// New creates and configures the application router
func New(h *handlers.Handler, staticDir string) chi.Router {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestLoggerMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.SessionMiddleware)

	r.Get("/health", healthCheck)

	r.Route("/api", func(r chi.Router) {
		r.Post("/markdown", h.CreateMarkdown)
		r.Get("/markdown/{id}", h.GetMarkdown)
		r.Delete("/markdown/{id}", h.DeleteMarkdown)
		r.Post("/markdown/{id}/comments", h.CreateComment)
		r.Get("/markdown/{id}/comments", h.GetComments)
		r.Get("/user/stats", h.GetUserStats)
	})

	r.Get("/view/{id}", serveViewPage(staticDir))
	r.Handle("/*", http.FileServer(http.Dir(staticDir)))

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func serveViewPage(staticDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/view.html")
	}
}
