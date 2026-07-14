package main 

import(
	"net/http"
)

func registerRoutes(mux *http.ServeMux,apiCfg *apiConfig){

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	))
	mux.HandleFunc("POST /api/refresh", apiCfg.HandlerRefresh)
	//create users
	mux.HandleFunc("POST /api/users", userHandler.CreateUser)

	mux.HandleFunc("POST /api/revoke", apiCfg.HandlerRevoke)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	//delete users
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirpByID)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	//a
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerPostPolkaWebhook)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGet)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetByID)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)

	//login
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

}