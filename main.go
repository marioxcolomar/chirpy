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
	polkaHandler := api.NewPolkaHandler(apiCfg.dbQueries, apiCfg.apiKey)
	userHandler := api.NewUserHandler(apiCfg.dbQueries, apiCfg.jwtSecret)
	authHandler := api.NewAuthHandler(apiCfg.dbQueries, apiCfg.jwtSecret)
	chirpHandler := api.NewChirpHandler(apiCfg.dbQueries, apiCfg.jwtSecret)
	adminHandler := api.NewAdminHandler(apiCfg.dbQueries, apiCfg.jwtSecret)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	mux.HandleFunc("POST /api/polka/webhooks", polkaHandler.HandleWebhooks)

	mux.HandleFunc("GET /api/healthz", handlerHealthCheck)

	mux.HandleFunc("POST /api/users", userHandler.HandleCreate)
	mux.HandleFunc("PUT /api/users", userHandler.HandleUpdate)

	mux.HandleFunc("POST /api/login", authHandler.HandleLogin)
	mux.HandleFunc("POST /api/refresh", authHandler.HandleTokenRefresh)
	mux.HandleFunc("POST /api/revoke", authHandler.HandleTokenRevoke)

	mux.HandleFunc("POST /api/chirps", chirpHandler.HandleCreate)
	mux.HandleFunc("GET /api/chirps", chirpHandler.HandleGet)
	mux.HandleFunc("GET /api/chirps/{chirpID}", chirpHandler.HandleGetOne)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", chirpHandler.HandleDelete)

	mux.HandleFunc("GET /admin/metrics", adminHandler.HandleMetrics)
	mux.HandleFunc("POST /admin/reset", adminHandler.HandleMetricsReset)

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
