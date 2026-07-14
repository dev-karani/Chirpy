package main

import (
	"net/http"

	"github.com/dev-karani/chirpy/chirps"
	"github.com/dev-karani/chirpy/users"
)

func registerRoutes(mux *http.ServeMux, apiCfg *apiConfig, users *users.Handler, chirps *chirps.Handler) {

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	))

	//user
	mux.HandleFunc("POST /api/refresh", users.Refresh)
	mux.HandleFunc("POST /api/users", users.CreateUser)
	mux.HandleFunc("POST /api/revoke", users.Revoke)
	mux.HandleFunc("PUT /api/users", users.UpdateUser)
	mux.HandleFunc("POST /api/login", users.Login)

    //chirps
	mux.HandleFunc("GET /api/chirps", chirps.GetChirp)
	mux.HandleFunc("GET /api/chirps/{chirpID}", chirps.GetChirpsByID)
	mux.HandleFunc("POST /api/chirps", chirps.CreateChirps)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", chirps.GetChirpsByID)
	
	//a
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerPostPolkaWebhook)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	//login
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

}
