package main

import (
	"fmt"
	"log"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/logger"
	"github.com/twsnmp/twsnmpfk/polling"
)

// MigrationProgress represents progress emitted during log migration.
type MigrationProgress struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Percent int    `json:"percent"`
	Bucket  string `json:"bucket"`
	Status  string `json:"status"` // "migrating", "done", "error"
	Message string `json:"message"`
}

// LogStoreInfo provides information about the current log store and migration status.
type LogStoreInfo struct {
	Format     string                  `json:"format"` // "parquet" or "bbolt"
	CanMigrate bool                    `json:"canMigrate"`
	Stats      datastore.BboltLogStats `json:"stats"`
}

// GetLogStoreInfo returns information about current log store.
func (a *App) GetLogStoreInfo() LogStoreInfo {
	stats := datastore.GetBboltLogStats()
	format := datastore.MapConf.LogFormat
	if stats.TotalCount > 0 {
		format = "bbolt"
		if datastore.MapConf.LogFormat != "bbolt" {
			datastore.MapConf.LogFormat = "bbolt"
			_ = datastore.SaveMapConf()
		}
	} else if format == "" {
		format = "bbolt"
	}
	canMigrate := format != "parquet" || stats.TotalCount > 0
	return LogStoreInfo{
		Format:     format,
		CanMigrate: canMigrate,
		Stats:      stats,
	}
}

// MigrateLogStoreToParquet executes migration from bbolt to Parquet.
func (a *App) MigrateLogStoreToParquet() error {
	log.Println("start MigrateLogStoreToParquet")

	// 1. Pause polling & logger
	polling.Pause()
	logger.Pause()
	defer func() {
		logger.Resume()
		polling.Resume()
	}()

	// 2. Report initial state
	wails.EventsEmit(a.ctx, "log-migration-progress", MigrationProgress{
		Current: 0,
		Total:   0,
		Percent: 0,
		Bucket:  "init",
		Status:  "migrating",
		Message: i18n.Trans("Preparing migration..."),
	})

	// 3. Run migration with progress callback
	err := datastore.MigrateBboltToParquet(func(bucket string, current, total int) {
		percent := 0
		if total > 0 {
			percent = (current * 100) / total
		}
		status := "migrating"
		msg := fmt.Sprintf("%s (%d / %d)", bucket, current, total)
		if bucket == "done" {
			status = "done"
			percent = 100
			msg = i18n.Trans("Migration completed")
		}
		wails.EventsEmit(a.ctx, "log-migration-progress", MigrationProgress{
			Current: current,
			Total:   total,
			Percent: percent,
			Bucket:  bucket,
			Status:  status,
			Message: msg,
		})
	})

	if err != nil {
		log.Printf("MigrateLogStoreToParquet err=%v", err)
		wails.EventsEmit(a.ctx, "log-migration-progress", MigrationProgress{
			Status:  "error",
			Message: err.Error(),
		})
		return err
	}

	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: i18n.Trans("Log storage migrated to Parquet"),
	})

	log.Println("completed MigrateLogStoreToParquet")
	return nil
}
