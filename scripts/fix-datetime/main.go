package main

import (
	"log"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/config"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/database"
)

const fixSQL = `
ALTER TABLE tasks
	ALTER COLUMN start_datetime  TYPE TIMESTAMPTZ USING NULLIF(start_datetime,'')::timestamptz,
	ALTER COLUMN end_datetime    TYPE TIMESTAMPTZ USING NULLIF(end_datetime,'')::timestamptz,
	ALTER COLUMN recurring_until TYPE TIMESTAMPTZ USING NULLIF(recurring_until,'')::timestamptz;
`

func main() {
	cfg := config.NewConfig()
	db := database.NewDB(cfg)
	log.Println("Running ALTER TABLE ...")
	if err := db.Exec(fixSQL).Error; err != nil {
		log.Fatalf("failed: %v", err)
	}
	log.Println("Done.")
}
