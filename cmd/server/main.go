package main

import (
	"log"
	"net/http"

	"github.com/AndreyBeloborodov/textlab/models"
)

func main() {
	modelRepository := models.NewMemoryRepository()
	modelService := models.NewService(modelRepository)
	modelHandler := models.NewHandler(modelService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/models", modelHandler.Create)
	mux.HandleFunc("GET /api/models", modelHandler.GetAll)
	mux.HandleFunc("GET /api/models/{id}", modelHandler.Get)
	mux.HandleFunc("PUT /api/models/{id}", modelHandler.Update)
	mux.HandleFunc("DELETE /api/models/{id}", modelHandler.Delete)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, "./web/index.html")
	})

	log.Println("Server started on http://localhost:8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}