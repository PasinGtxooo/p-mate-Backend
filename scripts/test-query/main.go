package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/config"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/database"
)

func main() {
	cfg := config.NewConfig()
	db := database.NewDB(cfg)

	today := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(2026, 11, 5, 23, 59, 59, 999999999, time.UTC)

	var tasks []*entity.Task
	err := db.WithContext(context.Background()).
		Where("start_datetime >= ? AND start_datetime <= ?", today, endOfDay).
		Find(&tasks).Error
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	fmt.Printf("ok, got %d tasks\n", len(tasks))
	for _, t := range tasks {
		var s string
		if t.StartDateTime != nil {
			s = *t.StartDateTime
		}
		fmt.Printf("  - %s  start=%q\n", t.Name, s)
	}
}
