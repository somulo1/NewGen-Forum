// Command seed loads the demo organization (the frontend's mock data and the tones) into an empty database.
// Development only: it refuses to run when HUB_ENV=prod.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/mail"
	"hub/internal/seeder"
)

func main() {
	deviceFor := flag.String("device-seed", "", "print the device seed of a demo person's email and exit")
	flag.Parse()
	secret := os.Getenv("HUB_SEED_SECRET")
	if secret == "" {
		secret = "hub-demo-seed"
	}
	if *deviceFor != "" {
		fmt.Println(seeder.DeviceSeed(secret, *deviceFor))
		return
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
	dev := &mail.Dev{Quiet: true, Cap: 5000}
	a, err := app.New(context.Background(), cfg, dev)
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
	defer a.Close()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	s := seeder.New(a, dev, secret)
	if err := s.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "seed failed:", err)
		os.Exit(1)
	}
	fmt.Printf("\nDemo organization ready.\n  sign in as %s / %s (org %q)\n", seeder.OwnerEmail, seeder.DemoPassword, seeder.OrgSlug)
}
