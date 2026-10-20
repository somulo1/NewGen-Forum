package chat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

// Likes, reactions and poll votes. There is no counter anywhere: every like or vote is its own message, written only
// through the API by a signed-in person, signed on their own device (the client drops anything whose signature does not
// check), sealed with a row MAC the database cannot forge, and counted by the reader one per person. These rules keep each
// person to one vote per poll and at most one of each reaction per post, and keep poll locks to the poll's author or a
// moderator.

const maxReactionsPerTarget = 2 // like and love

// checkEngagement runs inside Send for reactions, votes and poll locks.
func (s *Service) checkEngagement(ctx context.Context, tx pgx.Tx, p *tenant.Principal, conv uuid.UUID, kind, typ string, target *uuid.UUID) error {
	if typ != "reaction" && typ != "vote" && typ != "poll_lock" {
		return nil
	}
	if target == nil {
		return httpx.Bad("Say which post or poll this is for.")
	}
	var targetType string
	var author uuid.UUID
	var gone bool
	err := tx.QueryRow(ctx, `select content_type, sender_account, deleted_at is not null from hub.messages where org_id=$1 and conversation_id=$2 and id=$3`, p.Org, conv, *target).Scan(&targetType, &author, &gone)
	if errors.Is(err, pgx.ErrNoRows) || gone {
		return httpx.Bad("That post or poll is no longer there.")
	}
	if err != nil {
		return err
	}
	// one person's likes and votes on one target go one at a time, so two requests at once cannot both slip through
	if _, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtextextended('engage:' || $1::text || ':' || $2::text, 0))`, p.Account, *target); err != nil {
		return err
	}
	live := func(t string) (int, error) {
		var n int
		err := tx.QueryRow(ctx, `select count(*) from hub.messages where org_id=$1 and conversation_id=$2 and reply_to=$3 and sender_account=$4 and content_type=$5 and deleted_at is null`, p.Org, conv, *target, p.Account, t).Scan(&n)
		return n, err
	}
	switch typ {
	case "vote":
		if targetType != "poll" && targetType != "post" && targetType != "announcement" && targetType != "memo" {
			return httpx.Bad("You can only vote on a poll.")
		}
		n, err := live("vote")
		if err != nil {
			return err
		}
		if n > 0 {
			return httpx.Err(409, "already_voted", "You have already voted on this poll. Take your vote back first to change it.")
		}
	case "reaction":
		if targetType == "reaction" || targetType == "vote" || targetType == "poll_lock" {
			return httpx.Bad("You cannot react to that.")
		}
		n, err := live("reaction")
		if err != nil {
			return err
		}
		if n >= maxReactionsPerTarget {
			return httpx.Err(409, "already_reacted", "You have already reacted to this. Take your reaction back first.")
		}
	case "poll_lock":
		if targetType != "poll" {
			return httpx.Bad("Only a poll can be locked.")
		}
		if author != p.Account {
			if err := authz.Require(ctx, tx, p.Person, "forums", "Moderate"); err != nil {
				return httpx.Err(403, "forbidden", "Only the person who made the poll, or a moderator, can lock it.")
			}
		}
	}
	return nil
}

// reportTampering records messages whose row no longer matches its MAC (changed outside the system: a forged like or vote,
// an edited count, a rewritten poll). It writes to the audit log, which raises the "Tampered records" alert on the
// security logs page. Each message is reported once a day, however often it is read.
func reportTampering(ctx context.Context, a *app.App, org uuid.UUID, conv uuid.UUID, ids []uuid.UUID) {
	if len(ids) == 0 {
		return
	}
	fresh := ids[:0:0]
	for _, id := range ids {
		ok, err := a.Rdb.SetNX(ctx, "tamper:"+id.String(), 1, 24*time.Hour).Result()
		if err != nil || ok {
			fresh = append(fresh, id)
		}
	}
	if len(fresh) == 0 {
		return
	}
	shown := fresh
	if len(shown) > 8 {
		shown = shown[:8]
	}
	detail := fmt.Sprintf("Conversation %s: message(s) %v no longer match their seal (likes, votes or posts changed outside Hub). They are hidden from everyone. Restore them from a backup and find out who had direct access to the database.", conv, shown)
	err := a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		return a.AuditEntry(ctx, tx, &app.Entry{Org: org, Name: "Integrity monitor", Module: "Integrity", Result: "Failed",
			Action:  fmt.Sprintf("%d message(s) in a conversation were changed outside Hub (likes, votes or posts)", len(fresh)),
			Payload: map[string]any{"conversation": conv, "messages": fresh, "detail": detail}})
	})
	if err != nil {
		slog.Error("could not record tampering", "org", org, "err", err)
	}
}
