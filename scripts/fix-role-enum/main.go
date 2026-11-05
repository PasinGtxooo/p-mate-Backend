package main

import (
	"log"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/config"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/database"
)

func main() {
	cfg := config.NewConfig()
	db := database.NewDB(cfg)

	log.Println("Adding 'pm' to enum project_role (if missing) ...")
	if err := db.Exec(`ALTER TYPE project_role ADD VALUE IF NOT EXISTS 'pm'`).Error; err != nil {
		log.Fatalf("failed: %v", err)
	}
	log.Println("Done.")
}
