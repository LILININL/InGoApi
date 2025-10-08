package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"fristGoproject/internal/auth"
	"fristGoproject/internal/db"
	"fristGoproject/internal/httpapi"
	"fristGoproject/internal/user"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("unable to connect database: %v", err)
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	authSvc := auth.NewService(userRepo)
	authHandler := httpapi.NewAuthHandler(authSvc)
	userSvc := user.NewService(userRepo)
	userHandler := httpapi.NewUserHandler(userSvc)

	router := httpapi.NewRouter()
	router.RegisterAuthRoutes(authHandler)
	router.RegisterUserRoutes(userHandler)
	router.ServeDocs("docs")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router.Mux(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("server running at http://localhost:8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
