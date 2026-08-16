package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"products-manage-server/internal/config"
	infrapg "products-manage-server/internal/infrastructure/postgres"
	"products-manage-server/internal/product/application"
	"products-manage-server/internal/product/handler"
	productpg "products-manage-server/internal/product/infrastructure/postgres"
)

func main() {
	config.LoadDotEnv()
	databaseURL := os.Getenv("DATABASE_URL")
	db, err := infrapg.Open(databaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	repo := productpg.NewProductRepository(db)
	svc := application.NewProductService(repo)
	productHandler := handler.NewProductHandler(svc)

	r := mux.NewRouter()
	r.Use(corsMiddleware)
	productHandler.Register(r)

	addr := ":8080"
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Server started on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

// corsMiddleware はブラウザ（クライアント UI / Swagger UI）からのアクセスを許可する。
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" || origin == "http://localhost:8082" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
