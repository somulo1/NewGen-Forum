package seeder

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"

	"hub/internal/modules/chat"
	"hub/internal/modules/files"
)

func (s *Seeder) record(ctx context.Context, conv uuid.UUID, by *Actor, collection string, id uuid.UUID, payload any, meta map[string]any, parent *uuid.UUID) {
	in := chat.RecordIn{ID: id, Collection: collection, Parent: parent, Epoch: s.ConvEp[conv], Device: by.DevID, Envelope: s.sealFor(conv, by, id, payload), Meta: meta}
	_, err := s.Chat.SaveRecord(ctx, s.principal(by), conv, in, false)
	s.must(err, "record "+collection)
}

func (s *Seeder) seedNotes(ctx context.Context) {
	var books []struct{ ID, Name, Color string }
	var notes []struct {
		ID, Book, Title, Body string
		Updated               int64
		Pinned                bool
	}
	s.st("noteBooks", &books)
	s.st("notes", &notes)
	c, _ := s.newConversation(ctx, s.Owner, chat.CreateIn{Kind: "note"}, nil)
	ids := map[string]uuid.UUID{}
	for _, b := range books {
		id := idAt(s.ago(48 * time.Hour))
		ids[b.ID] = id
		s.record(ctx, c.ID, s.Owner, "notebook", id, map[string]any{"name": b.Name, "color": b.Color}, nil, nil)
	}
	for _, n := range notes {
		id := idAt(time.UnixMilli(n.Updated).UTC())
		meta := map[string]any{"pinned": n.Pinned}
		if bid, ok := ids[n.Book]; ok {
			meta["book"] = bid.String()
		}
		s.record(ctx, c.ID, s.Owner, "note", id, map[string]any{"title": n.Title, "body": n.Body, "updated": n.Updated}, meta, nil)
	}
}

type mockFile struct {
	Name, Kind, Size, By, When, ImgBg string
}

// encryptFile does what the app does before an upload: a fresh random file key, AES-256-GCM over the content.
func encryptFile(content []byte) (key, ciphertext []byte) {
	key = make([]byte, 32)
	_, _ = rand.Read(key)
	blk, _ := aes.NewCipher(key)
	g, _ := cipher.NewGCM(blk)
	nonce := make([]byte, g.NonceSize())
	_, _ = rand.Read(nonce)
	return key, g.Seal(nonce, nonce, content, nil)
}

func (s *Seeder) seedFiles(ctx context.Context) {
	var folders []struct {
		ID, Name, Color, Audience, Updated string
		Files                              []mockFile
	}
	s.st("folders", &folders)
	s.folder = map[string]uuid.UUID{}
	for _, f := range folders {
		in := chat.CreateIn{Kind: "folder", Name: f.Name, Color: f.Color}
		var members []*Actor
		switch f.Audience {
		case "everyone":
			in.Audience = "org"
			members = s.All[1:]
		case "Finance and managers":
			members = append(s.byDept("Finance"), s.byRole("hod")...)
		case "IT team":
			members = append(s.byDept("IT Operations"), s.byDept("Software Development")...)
		default: // project teams
			members = s.byRole("hod")
		}
		c, _ := s.newConversation(ctx, s.Owner, in, dedupeActors(members, s.Owner))
		s.folder[f.ID] = c.ID
		for _, mf := range f.Files {
			by := s.actor(mf.By)
			// the uploader must be in the folder to upload, so a folder always contains the people who put files in it
			if !s.inConversation(c.ID, by) {
				s.addToFolder(ctx, c.ID, by)
			}
			s.upload(ctx, c.ID, by, mf)
		}
	}
}

func (s *Seeder) upload(ctx context.Context, conv uuid.UUID, by *Actor, mf mockFile) uuid.UUID {
	content := bytes.Repeat([]byte("Hub demo file content for "+mf.Name+"\n"), 40+len(mf.Name))
	fileKey, ct := encryptFile(content)
	id := idAt(s.ago(parseAgoLoose(mf.When)))
	env := s.sealFor(conv, by, id, map[string]any{"name": mf.Name, "kind": mf.Kind, "size_label": mf.Size, "by": mf.By, "when": mf.When, "imgBg": mf.ImgBg, "file_key": base64.StdEncoding.EncodeToString(fileKey)})
	p := s.principal(by)
	_, err := s.File.Begin(ctx, p, conv, filesCreate(id, s.ConvEp[conv], by.DevID, env))
	s.must(err, "begin upload "+mf.Name)
	_, err = s.File.Upload(ctx, p, id, bytes.NewReader(ct), int64(len(ct)))
	s.must(err, "upload "+mf.Name)
	return id
}

// parseAgoLoose reads the mock's "2 days ago", "Last week", "Yesterday", "Last month".
func parseAgoLoose(s string) time.Duration {
	switch s {
	case "Last week":
		return 8 * 24 * time.Hour
	case "Last month":
		return 35 * 24 * time.Hour
	case "Yesterday":
		return 26 * time.Hour
	}
	return parseAgo(s)
}

func (s *Seeder) byDept(name string) []*Actor {
	id, ok := s.Units["department:"+name]
	if !ok {
		return nil
	}
	var out []*Actor
	for _, a := range s.All {
		if a.deptID == id {
			out = append(out, a)
		}
	}
	return out
}

func (s *Seeder) byRole(role string) []*Actor {
	var out []*Actor
	for _, a := range s.All {
		if a.Role == role {
			out = append(out, a)
		}
	}
	return out
}

func dedupeActors(l []*Actor, skip *Actor) []*Actor {
	seen := map[*Actor]bool{skip: true}
	var out []*Actor
	for _, a := range l {
		if !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	return out
}

func (s *Seeder) inConversation(conv uuid.UUID, a *Actor) bool {
	c, err := s.Chat.Get(context.Background(), s.principal(a), conv)
	return err == nil && c.MyRole != ""
}

// addToFolder lets the owner's device hand a new member the folder key, exactly as inviting someone does.
func (s *Seeder) addToFolder(ctx context.Context, conv uuid.UUID, a *Actor) {
	_, err := s.Chat.AddMembers(ctx, s.principal(s.Owner), conv, []uuid.UUID{a.Person}, s.keyWraps(conv, s.ConvKey[conv], s.ConvEp[conv], a), s.Owner.DevID)
	s.must(err, fmt.Sprintf("add %s to folder", a.Name))
}

func filesCreate(id uuid.UUID, epoch int, device uuid.UUID, env string) files.CreateIn {
	return files.CreateIn{ID: id, Epoch: epoch, Device: device, Envelope: env}
}
