package main

import (
	"log"
	"net/http"
	"pullrequests/internal/adapters/postgres"
	"pullrequests/internal/handlers"
	"pullrequests/internal/usecases"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sqlx.Open("postgres", "postgres://postgres:postgres@localhost:5432/pullrequests?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	teamRepo := postgres.NewTeamRepo(db)
	userRepo := postgres.NewUserRepo(db)
	pullrequestRepo := postgres.NewPRRepo(db)
	trm := postgres.NewSQLTransactionManager(db)

	teamUsecase := usecases.NewTeamUsecase(teamRepo, trm)
	userUsecase := usecases.NewUserUsecase(userRepo, trm)
	pullrequestUsecase := usecases.NewPRUsecase(pullrequestRepo, userRepo, trm)

	userHandler := handlers.NewUserHandler(userUsecase)
	teamHandler := handlers.NewTeamHandler(teamUsecase)
	pullrequestHandler := handlers.NewPRHandler(pullrequestUsecase)

	r := chi.NewRouter()

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.AddTeam)
		r.Get("/get", teamHandler.GetTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetUserActive)
		r.Get("/getReview", pullrequestHandler.GetUserReviewPRs)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", pullrequestHandler.CreatePR)
		r.Post("/merge", pullrequestHandler.MergePR)
		r.Post("/reassign", pullrequestHandler.ReassignReviewer)
	})

	port := ":8080"
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
