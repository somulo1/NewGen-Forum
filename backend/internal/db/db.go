// Package db opens the PostgreSQL pool and runs work inside tenant-scoped transactions.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct{ Pool *pgxpool.Pool }

// Open connects as the runtime role. Every new connection proves it is the Hub system by setting the guard token,
// which the database checks on every write (see hub.assert_app). Without it, inserts/updates/deletes are refused.
func Open(ctx context.Context, url, guardToken string, maxConns int32) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	if maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	cfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, "select set_config('hub.app_token', $1, false)", guardToken)
		return err
	}
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := p.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database not reachable: %w", err)
	}
	return &DB{Pool: p}, nil
}

func (d *DB) Close() { d.Pool.Close() }

// Tx runs fn in a transaction with no organization (global tables such as accounts).
func (d *DB) Tx(ctx context.Context, fn func(pgx.Tx) error) error { return d.run(ctx, uuid.Nil, fn) }

// WithOrg runs fn in a transaction scoped to one organization. Row level security then makes every other
// organization's rows invisible and unwritable, even if a query forgets its WHERE clause.
func (d *DB) WithOrg(ctx context.Context, org uuid.UUID, fn func(pgx.Tx) error) error {
	if org == uuid.Nil {
		return errors.New("db: organization required")
	}
	return d.run(ctx, org, fn)
}

func (d *DB) run(ctx context.Context, org uuid.UUID, fn func(pgx.Tx) error) error {
	opts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	if org != uuid.Nil {
		// begin and set the organization in one trip to the database (a uuid is only hex digits and dashes, so it cannot carry SQL)
		opts.BeginQuery = "begin isolation level read committed; select set_config('hub.org_id', '" + org.String() + "', true)"
	}
	tx, err := d.Pool.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SetOrg scopes the rest of an open transaction to one organization (used when the organization is created inside it).
func SetOrg(ctx context.Context, tx pgx.Tx, org uuid.UUID) error {
	_, err := tx.Exec(ctx, "select set_config('hub.org_id', $1, true)", org.String())
	return err
}
