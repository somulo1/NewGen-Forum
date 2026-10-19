// Command migrate applies database migrations with the owner role.
package main

import (
	"context"
	"fmt"
	"os"

	"hub/internal/config"
	"hub/internal/db"
	"hub/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}
	applied, err := db.Migrate(context.Background(), cfg.MigrateURL, cfg.DBGuard, migrations.FS)
	if err != nil {
		fatal(err)
	}
	if len(applied) == 0 {
		fmt.Println("database is up to date")
		return
	}
	for _, a := range applied {
		fmt.Println("applied", a)
	}
}

func fatal(err error) { fmt.Fprintln(os.Stderr, "migrate:", err); os.Exit(1) }
