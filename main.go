package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/marioxcolomar/chirpy/api"
	"github.com/marioxcolomar/chirpy/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	// DB
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Unable to open connection to database: \n", err)
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("."))
	handler := http.StripPrefix("/app", fs)

	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	apiKey := os.Getenv("POLKA_KEY")

	apiCfg := apiConfig{
		dbQueries: dbQueries,
		platform:  platform,
		jwtSecret: jwtSecret,
		apiKey:    apiKey,
	}

	// Handler
	userHandler := api.NewUserHandler(apiCfg.dbQueries, apiCfg.jwtSecret)
	chirpHandler := api.NewChirpHandler(apiCfg.dbQueries, apiCfg.jwtSecret)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerWebhooks)

	mux.HandleFunc("GET /api/healthz", handlerHealthCheck)

	mux.HandleFunc("POST /api/users", userHandler.HandleCreate)
	mux.HandleFunc("PUT /api/users", userHandler.HandleUpdate)

	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerTokenRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerTokenRevoke)

	mux.HandleFunc("POST /api/chirps", chirpHandler.HandleCreate)
	mux.HandleFunc("GET /api/chirps", chirpHandler.HandleGet)
	mux.HandleFunc("GET /api/chirps/{chirpID}", chirpHandler.HandleGetOne)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", chirpHandler.HandleDelete)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerMetricsReset)

	const port = "8080"

	srv := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	fmt.Printf("server started on: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	jwtSecret      string
	apiKey         string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("incoming request for route: %s\n", r.URL)
		next.ServeHTTP(w, r)
	})
}
