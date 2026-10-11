package main

import (
	"log"
	"net/http"
	"time"

	"github.com/AndreyBeloborodov/textlab/auth"
	"github.com/AndreyBeloborodov/textlab/models"
	"github.com/AndreyBeloborodov/textlab/users"
    "github.com/AndreyBeloborodov/textlab/algorithm"
    "github.com/AndreyBeloborodov/textlab/training"
    "github.com/AndreyBeloborodov/textlab/prediction"
)

func main() {
	secret := "local-development-secret-at-least-32-chars"

	accessTokenTTL := 15 * time.Minute
	refreshTokenTTL := 30 * 24 * time.Hour

	userRepository := users.NewMemoryRepository()
	userService := users.NewService(userRepository)
	userHandler := users.NewHandler(userService)

	tokenManager, err := auth.NewJWTManager(secret, accessTokenTTL)
	if err != nil {
		log.Fatalf("create token manager: %v", err)
	}

	refreshRepository := auth.NewMemoryRefreshRepository()
	passwordHasher := auth.NewArgon2PasswordHasher()

	authService := auth.NewService(
    	userService,
    	passwordHasher,
    	tokenManager,
    	refreshRepository,
    	refreshTokenTTL,
    )
	authHandler := auth.NewHandler(
		authService,
		refreshTokenTTL,
		false, // Для локальной разработки без HTTPS.
	)

	authMiddleware := auth.NewMiddleware(tokenManager)

	modelRepository := models.NewMemoryRepository()
	modelService := models.NewService(modelRepository)
	modelHandler := models.NewHandler(modelService)

    trainingRepository := training.NewMemoryRepository()

    store := algorithm.NewMemoryStore()

    trainingManager := training.NewManager(
        trainingRepository,
        algorithm.NewHashAlgorithm,
        store,
    )
    
    trainingService := training.NewService(
        modelService,
        trainingRepository,
        trainingManager,
    )
    
    trainingHandler := training.NewHandler(trainingService)

    predictionService := prediction.NewService(
    	modelService,
    	store,
    	algorithm.NewHashAlgorithm,
    )
    
    predictionHandler := prediction.NewHandler(predictionService)

	mux := http.NewServeMux()

	// Авторизация: эти маршруты не требуют access-токена.
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)

	// CRUD пользователей.
	// В этом примере маршруты доступны без авторизации.
	// Если они должны быть приватными, оберни их в authMiddleware.
	mux.HandleFunc("GET /api/users", userHandler.GetAll)
	mux.HandleFunc("GET /api/users/{id}", userHandler.Get)
	mux.HandleFunc("PUT /api/users/{id}", userHandler.Update)
	mux.HandleFunc("DELETE /api/users/{id}", userHandler.Delete)

	// Операции с моделями требуют валидного access-токена.
	mux.Handle("POST /api/models",
		authMiddleware.Authenticate(http.HandlerFunc(modelHandler.Create)),
	)
	mux.Handle("GET /api/models",
		authMiddleware.Authenticate(http.HandlerFunc(modelHandler.GetAll)),
	)
	mux.Handle("GET /api/models/{id}",
		authMiddleware.Authenticate(http.HandlerFunc(modelHandler.Get)),
	)
	mux.Handle("PUT /api/models/{id}",
		authMiddleware.Authenticate(http.HandlerFunc(modelHandler.Update)),
	)
	mux.Handle("DELETE /api/models/{id}",
		authMiddleware.Authenticate(http.HandlerFunc(modelHandler.Delete)),
	)

    mux.Handle("POST /api/models/{id}/train",
        authMiddleware.Authenticate(http.HandlerFunc(trainingHandler.Start)),
    )
    mux.Handle("GET /api/trainings/{id}",
        authMiddleware.Authenticate(http.HandlerFunc(trainingHandler.GetStatus)),
    )

    mux.Handle("POST /api/models/{id}/predict",
        authMiddleware.Authenticate(http.HandlerFunc(predictionHandler.Predict)),
    )

	log.Println("TextLab server listening on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
