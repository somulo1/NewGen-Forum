package chat_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/e2ee"
	"hub/internal/modules/chat"
	"hub/internal/testkit"
)

func putBytes(t *testing.T, e *testkit.Env, method, path string, body []byte, token string) (int, []byte) {
	t.Helper()
	req, _ := http.NewRequest(method, e.Srv.URL+"/v1"+path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Forwarded-For", "10.8."+uuid.NewString()[:2]+".9")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, b
}

// A deleted post takes its pictures with it at once, and a moment past its 24 hours is erased for good by the sweep, files included.
func TestDeletedPostsAndExpiredMomentsLeaveNoFilesBehind(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	o.Register(e)
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Team", "members": []string{o.Person}, "keys": o.WrapFor(e, key, 1, o.Person)}, ownerU.Access)["id"].(string)
	chats := chat.New(e.A)

	upload := func() string {
		fid := uuid.Must(uuid.NewV7()).String()
		e.MustDo(201, "POST", "/folders/"+conv+"/files", map[string]any{"id": fid, "epoch": 1, "device": o.DevID, "envelope": o.Seal(org, conv, fid, 1, key, `{"name":"pic.png"}`)}, o.Access)
		if code, _ := putBytes(t, e, "PUT", "/files/"+fid+"/content", []byte("sealed picture bytes"), o.Access); code != 200 {
			t.Fatalf("upload: %d", code)
		}
		return fid
	}
	send := func(typ, fid string) string {
		id := uuid.Must(uuid.NewV7()).String()
		e.MustDo(201, "POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "type": typ, "epoch": 1, "device": o.DevID, "envelope": o.Seal(org, conv, id, 1, key, "x"), "files": []string{fid}}, o.Access)
		return id
	}
	status := func(fid string) string {
		var st string
		_ = e.A.DB.WithOrg(context.Background(), uuid.MustParse(org), func(tx pgx.Tx) error {
			return tx.QueryRow(context.Background(), `select status from hub.files where id=$1`, fid).Scan(&st)
		})
		return st
	}
	gone := func(fid string) bool {
		code, _ := putBytes(t, e, "GET", "/files/"+fid+"/content", nil, o.Access)
		return code == 404 && status(fid) == "deleted"
	}
	// before: the file is there; deleting the post removes it at once (no waiting for the worker)
	f1 := upload()
	post := send("post", f1)
	if code, _ := putBytes(t, e, "GET", "/files/"+f1+"/content", nil, o.Access); code != 200 {
		t.Fatalf("file unreadable before delete: %d", code)
	}
	e.MustDo(204, "DELETE", "/conversations/"+conv+"/messages/"+post, nil, o.Access)
	if !gone(f1) {
		t.Fatal("the post's file is still there after the post was deleted")
	}
	// a fresh moment stays; one older than 24 hours is hidden, then erased by the sweep
	f2 := upload()
	moment := send("moment", f2)
	if n, err := chats.ExpireMoments(context.Background()); err != nil || status(f2) == "deleted" {
		t.Fatalf("a fresh moment was erased: n=%d err=%v", n, err)
	}
	root, err := pgx.Connect(context.Background(), "postgres://hubadmin@/hub?host=/tmp&port=5433")
	if err != nil {
		t.Skip("superuser not reachable: ", err)
	}
	defer root.Close(context.Background())
	_, _ = root.Exec(context.Background(), "set session_replication_role = replica")
	if _, err := root.Exec(context.Background(), `update hub.messages set created_at = now() - interval '25 hours' where id=$1`, moment); err != nil {
		t.Fatal(err)
	}
	for _, m := range e.MustDo(200, "GET", "/conversations/"+conv+"/messages", nil, o.Access)["messages"].([]any) {
		if m.(map[string]any)["id"] == moment {
			t.Fatal("an expired moment is still listed")
		}
	}
	if n, err := chats.ExpireMoments(context.Background()); err != nil || n < 1 {
		t.Fatalf("sweep: n=%d err=%v", n, err)
	}
	if !gone(f2) {
		t.Fatal("the moment's file is still there after the sweep")
	}
}

// A device that was offline decides what to catch up on from changed_at: it must move for a message, a like and a removal.
func TestConversationChangedAtMovesForEveryKindOfChange(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	o.Register(e)
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Team", "members": []string{o.Person}, "keys": o.WrapFor(e, key, 1, o.Person)}, ownerU.Access)["id"].(string)
	changed := func() string {
		t.Helper()
		v, _ := e.MustDo(200, "GET", "/conversations/"+conv, nil, o.Access)["changed_at"].(string)
		if v == "" {
			t.Fatal("changed_at is missing")
		}
		return v
	}
	last := changed()
	step := func(what string) {
		t.Helper()
		now := changed()
		if now <= last {
			t.Fatalf("changed_at did not move for %s: %s -> %s", what, last, now)
		}
		last = now
	}
	send := func(typ string, replyTo any) string {
		id := uuid.Must(uuid.NewV7()).String()
		e.MustDo(201, "POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "type": typ, "reply_to": replyTo, "epoch": 1, "device": o.DevID, "envelope": o.Seal(org, conv, id, 1, key, "x")}, o.Access)
		return id
	}
	msg := send("text", nil)
	step("a message")
	like := send("reaction", msg)
	step("a like")
	e.MustDo(204, "DELETE", "/conversations/"+conv+"/messages/"+like, nil, o.Access)
	step("a removed like")
	e.MustDo(204, "DELETE", "/conversations/"+conv+"/messages/"+msg, nil, o.Access)
	step("a removed message")
}

// Deleting a post erases everything that hangs on it, however deep: comments, replies, likes and votes.
func TestDeletingAPostErasesItsWholeThread(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	o.Register(e)
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Team", "members": []string{o.Person}, "keys": o.WrapFor(e, key, 1, o.Person)}, ownerU.Access)["id"].(string)
	send := func(typ string, replyTo any) string {
		id := uuid.Must(uuid.NewV7()).String()
		e.MustDo(201, "POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "type": typ, "reply_to": replyTo, "epoch": 1, "device": o.DevID, "envelope": o.Seal(org, conv, id, 1, key, "x")}, o.Access)
		return id
	}
	post := send("post", nil)
	comment := send("comment", post)
	reply := send("comment", comment)
	send("reaction", reply)
	send("reaction", post)
	send("vote", post)
	live := func() int {
		var n int
		_ = e.A.DB.WithOrg(context.Background(), uuid.MustParse(org), func(tx pgx.Tx) error {
			return tx.QueryRow(context.Background(), `select count(*) from hub.messages where conversation_id=$1 and deleted_at is null`, conv).Scan(&n)
		})
		return n
	}
	if live() != 6 {
		t.Fatalf("setup: want 6 live messages, got %d", live())
	}
	e.MustDo(204, "DELETE", "/conversations/"+conv+"/messages/"+post, nil, o.Access)
	if n := live(); n != 0 {
		t.Fatalf("%d messages of the deleted post's thread are still there", n)
	}
}
