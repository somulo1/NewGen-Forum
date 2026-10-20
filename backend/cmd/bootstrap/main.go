// Command bootstrap installs the reference data a live deployment needs and
// never creates people, messages or any sample content.
//
//	bootstrap platform                   install the shipped sounds (safe to run again; the API also does this at start-up)
//	bootstrap protect [-slug acme]       check the encryption and database protections are in place; with -slug also make sure that
//	                                     organization has its data key. Exits non-zero if anything is wrong.
//	bootstrap org -slug acme             make sure the organization has its catalog of pages and rights
//	bootstrap org -slug acme -roles starter   ...and add the starter roles (Admin, Head of department, Staff, ...)
//	bootstrap org -slug acme -roles starter -dry-run    say what would be added, change nothing
//
// Everything is idempotent: running it again adds only what is missing and never overwrites what an organization changed.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/mail"
	"hub/internal/modules/tones"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "bootstrap:", err)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: bootstrap platform | bootstrap protect [-slug <slug>] | bootstrap org -slug <slug> [-roles starter] [-dry-run]")
		os.Exit(2)
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	slug := fs.String("slug", "", "organization web name")
	roles := fs.String("roles", "", "role pack to add: starter")
	all := fs.Bool("all", false, "every organization (use after a release that adds rights)")
	dry := fs.Bool("dry-run", false, "show what would be added and change nothing")
	_ = fs.Parse(os.Args[2:])

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	a, err := app.New(context.Background(), cfg, &mail.Dev{Quiet: true})
	if err != nil {
		fail(err)
	}
	defer a.Close()
	ctx := context.Background()

	switch cmd {
	case "platform":
		if err := tones.New(a).Install(ctx); err != nil {
			fail(err)
		}
		fmt.Println("platform reference data installed (sounds).")
	case "org":
		if *slug == "" && !*all {
			fail(fmt.Errorf("-slug or -all is required"))
		}
		if *roles != "" && *roles != "starter" {
			fail(fmt.Errorf("unknown role pack %q (available: starter)", *roles))
		}
		var orgs []uuid.UUID
		if err := a.DB.Tx(ctx, func(tx pgx.Tx) error {
			q, args := `select id from hub.orgs where slug=$1`, []any{*slug}
			if *all {
				q, args = `select id from hub.orgs where status='active' order by created_at`, nil
			}
			rs, err := tx.Query(ctx, q, args...)
			if err != nil {
				return err
			}
			defer rs.Close()
			for rs.Next() {
				var id uuid.UUID
				if err := rs.Scan(&id); err != nil {
					return err
				}
				orgs = append(orgs, id)
			}
			return rs.Err()
		}); err != nil || len(orgs) == 0 {
			fail(fmt.Errorf("no organization found for %q", *slug))
		}
		for _, org := range orgs {
			fmt.Println("organization", org)
			seal := func(tx pgx.Tx) func(table, where string, args ...any) error {
				return func(table, where string, args ...any) error { return a.Sign.Seal(ctx, tx, table, where, args...) }
			}
			err := a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
				if err := access.Install(ctx, tx, org, seal(tx)); err != nil {
					return err
				}
				fmt.Println("catalog of pages and rights: present")
				if *roles == "starter" {
					ids, err := access.InstallStarterRoles(ctx, tx, org, app.NewID, seal(tx))
					if err != nil {
						return err
					}
					fmt.Printf("starter roles added: %d (existing roles were left as they are)\n", len(ids))
				}
				if *dry {
					return fmt.Errorf("dry run: rolled back")
				}
				return nil
			})
			if err != nil {
				if *dry && err.Error() == "dry run: rolled back" {
					fmt.Println("dry run: nothing was changed.")
					continue
				}
				fail(err)
			}
		}
		fmt.Println("done.")
	case "protect":
		os.Exit(protect(ctx, a, cfg, *slug))
	default:
		fail(fmt.Errorf("unknown command %q", cmd))
	}
}
