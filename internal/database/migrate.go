package database

import (
	"log"

	"github.com/jmoiron/sqlx"
)

func AutoMigrate(db *sqlx.DB, schema string) error {
	log.Println("[DB] Running auto-migration...")
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	log.Println("[DB] Auto-migration completed successfully")
	return nil
}
