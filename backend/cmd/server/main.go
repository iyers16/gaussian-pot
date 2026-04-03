package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/iyers16/gaussian-pot/backend/internal/game"
	"github.com/iyers16/gaussian-pot/backend/internal/handler"
	"github.com/iyers16/gaussian-pot/backend/internal/repository"
)

func main() {
	db := mustConnectDB()
	mustMigrate(db)

	// Repositories.
	userRepo := repository.NewUserRepository(db)
	roundRepo := repository.NewRoundRepository(db)
	betRepo := repository.NewBetRepository(db)
	debtRepo := repository.NewDebtRepository(db)

	// Game primitives.
	sm := game.NewStateMachine()
	hub := handler.NewHub()
	sessions := handler.NewSessionStore()

	// Handlers.
	authH := handler.NewAuthHandler(userRepo, sessions, hub)
	roundH := handler.NewRoundHandler(betRepo, debtRepo, userRepo, roundRepo, sm, hub)
	adminH := handler.NewAdminHandler(roundRepo, betRepo, debtRepo, userRepo, sm, hub)
	debtH := handler.NewDebtHandler(debtRepo, roundRepo, sm, hub)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// WebSocket — no auth required (client sends token after connecting if needed).
	r.GET("/ws", hub.HandleWS)

	api := r.Group("/api")
	{
		// Public auth.
		api.POST("/auth/login", authH.Login)
		api.POST("/auth/logout", handler.AuthMiddleware(sessions), authH.Logout)

		// Authenticated player routes.
		auth := api.Group("/", handler.AuthMiddleware(sessions))
		{
			auth.GET("/round", roundH.GetCurrentRound)
			auth.POST("/round/bet", roundH.PlaceBet)
			auth.GET("/debts", debtH.GetMyDebts)
			auth.POST("/debts/:id/confirm-paid", debtH.ConfirmPaid)
			auth.POST("/debts/:id/confirm-received", debtH.ConfirmReceived)
		}

		// Host-only admin routes.
		admin := api.Group("/admin", handler.AuthMiddleware(sessions), handler.HostMiddleware())
		{
			admin.POST("/round/open", adminH.OpenRound)
			admin.POST("/round/call", adminH.CallRound)
			admin.POST("/credits/replenish", adminH.ReplenishCredits)
			admin.POST("/question/random", adminH.RandomQuestion)
			admin.GET("/questions", adminH.ListQuestions)
			admin.GET("/users", adminH.ListUsers)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on :%s", port)
	r.Run("0.0.0.0:" + port)
}

func mustConnectDB() *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func mustMigrate(db *sql.DB) {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS numbers (
			id SERIAL PRIMARY KEY,
			value INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			role VARCHAR(10) NOT NULL DEFAULT 'player',
			credits NUMERIC(10,2) NOT NULL DEFAULT 1000.00,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS rounds (
			id SERIAL PRIMARY KEY,
			question_id INTEGER NOT NULL,
			question_text TEXT NOT NULL,
			target_value NUMERIC(10,2) NOT NULL,
			unit VARCHAR(50) NOT NULL,
			mode VARCHAR(20) NOT NULL,
			state VARCHAR(20) NOT NULL DEFAULT 'ROUND_OPEN',
			opened_at TIMESTAMP DEFAULT NOW(),
			called_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS bets (
			id SERIAL PRIMARY KEY,
			round_id INTEGER REFERENCES rounds(id),
			user_id INTEGER REFERENCES users(id),
			username VARCHAR(50) NOT NULL,
			guess NUMERIC(10,2) NOT NULL,
			wager NUMERIC(10,2) NOT NULL,
			payout NUMERIC(10,2) DEFAULT 0,
			placed_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(round_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS debts (
			id SERIAL PRIMARY KEY,
			round_id INTEGER REFERENCES rounds(id),
			payer_id INTEGER REFERENCES users(id),
			payee_id INTEGER REFERENCES users(id),
			payer_username VARCHAR(50) NOT NULL,
			payee_username VARCHAR(50) NOT NULL,
			amount NUMERIC(10,2) NOT NULL,
			payer_confirmed BOOLEAN DEFAULT FALSE,
			payee_confirmed BOOLEAN DEFAULT FALSE,
			status VARCHAR(20) DEFAULT 'PENDING',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
	}
	log.Println("DB migrations complete")
}
