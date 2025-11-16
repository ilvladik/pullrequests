package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pullrequests/internal/adapters/postgres"
	"pullrequests/internal/config"
	"pullrequests/internal/handlers"
	"pullrequests/internal/usecases"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	db, err := sqlx.Open("postgres", cfg.GetConnectionString())
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("db unreachable: %v", err)
	}

	teamRepo := postgres.NewTeamRepo(db)
	userRepo := postgres.NewUserRepo(db)
	pullrequestRepo := postgres.NewPRRepo(db)
	trm := postgres.NewSQLTransactionManager(db)

	teamUsecase := usecases.NewTeamUsecase(teamRepo, userRepo, trm)
	userUsecase := usecases.NewUserUsecase(userRepo, trm)
	pullrequestUsecase := usecases.NewPRUsecase(pullrequestRepo, userRepo, trm)

	teamHandler := handlers.NewTeamHandler(teamUsecase)
	userHandler := handlers.NewUserHandler(userUsecase)
	pullRequestHandler := handlers.NewPRHandler(pullrequestUsecase)

	r := chi.NewRouter()

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.AddTeam)
		r.Get("/get", teamHandler.GetTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetUserActive)
		r.Get("/getReview", pullRequestHandler.GetUserReviewPRs)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", pullRequestHandler.CreatePR)
		r.Post("/merge", pullRequestHandler.MergePR)
		r.Post("/reassign", pullRequestHandler.ReassignReviewer)
	})

	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("server started on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown: %v", err)
	}

	<-ctx.Done()
	log.Println("Shutdown timeout reached")
}
