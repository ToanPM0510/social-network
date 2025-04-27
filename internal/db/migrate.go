package db

import (
	"database/sql"
	"os"
)

func RunMigration(db *sql.DB, filepath string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(content))
	return err
}
