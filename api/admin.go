package api

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/marioxcolomar/chirpy/internal/database"
)

type AdminHandler struct {
	db             *database.Queries
	fileserverHits *atomic.Int32
	platform       string
}

func NewAdminHandler(db *database.Queries, fileserverHits *atomic.Int32, platform string) *AdminHandler {
	return &AdminHandler{db: db, fileserverHits: fileserverHits, platform: platform}
}

func (h *AdminHandler) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.fileserverHits.Add(1)
		fmt.Printf("incoming request for route: %s\n", r.URL)
		next.ServeHTTP(w, r)
	})
}
func (h *AdminHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	page := fmt.Sprintf(`
<html>
<body>
<h1>Welcome, Chirpy Admin</h1>
<p>Chirpy has been visited %d times!</p>
</body>
</html>
`,
		h.fileserverHits.Load())
	w.Write([]byte(page))
}

func (h *AdminHandler) HandleMetricsReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if h.platform != "dev" {
		w.WriteHeader(401)
		return
	}
	w.WriteHeader(200)
	h.fileserverHits.Store(0)
	h.db.DeleteUsers(r.Context())
}
