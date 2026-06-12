package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
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
		cfg.fileserverHits.Load())
	w.Write([]byte(page))
}

func (cfg *apiConfig) handleMetricsReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if cfg.platform != "dev" {
		w.WriteHeader(401)
		return
	}
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	cfg.dbQueries.DeleteUsers(r.Context())
}
