package seeder

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/e2ee"
	"hub/internal/modules/auth"
	"hub/internal/modules/keys"
	"hub/internal/modules/org"
	"hub/internal/modules/people"
	"hub/internal/modules/roles"
	"hub/internal/modules/tones"
	"hub/seed"
)

func (s *Seeder) seedTones(ctx context.Context) {
	b, err := seed.FS.ReadFile("tones.json")
	s.must(err, "tones file")
	var list []tones.Tone
	s.must(json.Unmarshal(b, &list), "tones json")
	s.must(s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		for _, t := range list {
			if err := s.Tone.Put(ctx, tx, t); err != nil {
				return err
			}
		}
		return nil
	}), "store tones")
}

type orgData struct {
	Sectors  []struct{ ID, Name string } `json:"sectors"`
	Branches []struct {
		ID, Name, Sector string
	} `json:"branches"`
	Units []struct {
		ID, Name, Branch string
	} `json:"units"`
	Departments []struct {
		ID, Name, Branch, Sector string
	} `json:"departments"`
	Mine struct{ Departments, Branches, Sectors, Units string } `json:"mine"`
}

type baseLog struct {
	ID     string  `json:"id"`
	Ago    float64 `json:"ago"`
	User   string  `json:"user"`
	Action string  `json:"action"`
	Module string  `json:"module"`
	IP     string  `json:"ip"`
	Loc    string  `json:"loc"`
	Result string  `json:"result"`
	Rule   *string `json:"rule"`
}

// namedRoles are the roles the frontend's mock role pages show for a few well-known people.
var namedRoles = map[string]string{
	"David Otieno": "admin", "Grace Njeri": "admin", "Amina Wanjiru": "hod", "Kevin Mutua": "hod",
	"Brian Kiprop": "it", "John Kamau": "it", "Mercy Wambui": "finance",
}

var headRe = regexp.MustCompile(`(?i)^(head of|director|chief)`)

func roleFor(d dirPerson) string {
	if r, ok := namedRoles[d.Name]; ok {
		return r
	}
	switch {
	case headRe.MatchString(d.Title):
		return "hod"
	case strings.Contains(d.Title, "Sector"):
		return "sector"
	case strings.Contains(d.Title, "Branch Manager"):
		return "branch"
	case strings.Contains(d.Title, "Unit Lead"):
		return "unit"
	}
	return "staff"
}

func (s *Seeder) seedPeople(ctx context.Context) {
	quiz := []auth.QuizIn{{Question: "First pet", Answer: "Rex"}, {Question: "Home town", Answer: "Nairobi"}, {Question: "First school", Answer: "Hub Primary"}}
	s.A.PW = fastPW // sign-up hashes passwords and answers; demo accounts do not need production-grade cost
	// the owner
	acct, err := s.Auth.Signup(ctx, OwnerName, OwnerEmail, DemoPassword, quiz, "10.0.0.1")
	s.must(err, "owner signup")
	o := &Actor{Name: OwnerName, First: "Samwel", Last: "Okoth Omulo", Email: OwnerEmail, Account: acct, Color: "#B8F36B", Initials: "S", Role: "superadmin"}
	created, err := s.Org.Create(ctx, acct, OwnerName, "Demo Organization", OrgSlug, "10.0.0.1")
	s.must(err, "create organization")
	s.OrgID, o.Person = created.ID, created.Person
	o.P = s.principal(o)
	s.Owner = o
	s.All = append(s.All, o)
	s.ByEmail[OwnerEmail], s.ByName[OwnerName] = o, o
	s.must(s.deviceUp(ctx, o), "owner device")

	// Security rules stay quiet while the demo is being built (it is built at any hour, by one administrator);
	// seedSecurity switches them on before the history is replayed.
	for _, r := range []string{"r4", "r5"} {
		s.must(s.Sec.SetRule(ctx, o.P, r, false, "Medium"), "quiet rule "+r)
	}
	// structure
	add := func(kind, name string, parent *uuid.UUID, mock string) uuid.UUID {
		u, err := s.Org.SaveUnit(ctx, o.P, nil, org.Unit{Kind: kind, Name: name, Parent: parent})
		s.must(err, "unit "+name)
		s.Units[kind+":"+name], s.UnitIDs[mock] = u.ID, u.ID
		return u.ID
	}
	for _, x := range s.data.org.Sectors {
		add("sector", x.Name, nil, x.ID)
	}
	for _, x := range s.data.org.Branches {
		p := s.UnitIDs[x.Sector]
		add("branch", x.Name, &p, x.ID)
	}
	for _, x := range s.data.org.Units {
		p := s.UnitIDs[x.Branch]
		add("unit", x.Name, &p, x.ID)
	}
	for _, x := range s.data.org.Departments {
		p := s.UnitIDs[x.Branch]
		add("department", x.Name, &p, x.ID)
	}

	// People. Roles that need approval to change are set up directly now; the flow is restored afterwards.
	people := make([]people.NewPerson, 0, len(s.data.dir)+1)
	type pending struct {
		d dirPerson
		a *Actor
	}
	var all []pending
	unit := func(kind, name string) *uuid.UUID {
		if id, ok := s.Units[kind+":"+name]; ok && name != "" {
			return &id
		}
		return nil
	}
	for _, d := range s.data.dir {
		first, last, _ := strings.Cut(d.Name, " ")
		a := &Actor{Name: d.Name, First: first, Last: last, Email: d.Email, Title: d.Title, Color: d.Color, Initials: d.Initials, Availability: d.Status, Role: roleFor(d)}
		if id := unit("department", d.Dept); id != nil {
			a.deptID = *id
		}
		all = append(all, pending{d, a})
		people = append(people, nPerson(first, last, d, a, unit))
	}
	// one guest from outside, as in the mock role page
	guest := &Actor{Name: "External auditor", First: "External", Last: "auditor", Email: "auditor@external.example", Title: "Auditor", Color: "#9DB0A6", Initials: "EA", Role: "guest", Availability: "Available"}
	all = append(all, pending{dirPerson{Name: guest.Name, Email: guest.Email, Title: "Auditor"}, guest})
	people = append(people, nPersonRaw(guest))

	for i := 0; i < len(people); i += 100 {
		j := min(i+100, len(people))
		res, err := s.Ppl.Onboard(ctx, o.P, people[i:j])
		s.must(err, "onboard")
		for k, r := range res {
			if r.Status != "invited" {
				panic(fmt.Sprintf("seed: could not add %s: %s", people[i+k].Email, r.Error))
			}
			all[i+k].a.Person = r.ID
		}
	}
	// every person signs up, accepts the emailed invitation and registers a device
	for _, pd := range all {
		a := pd.a
		acct, err := s.Auth.Signup(ctx, a.Name, a.Email, DemoPassword, quiz, "10.0.0.1")
		s.must(err, "signup "+a.Email)
		a.Account = acct
		msgs := s.Dev.To(a.Email)
		if len(msgs) == 0 {
			panic("seed: no invitation was mailed to " + a.Email)
		}
		m := regexp.MustCompile(`/join\?t=(\S+)`).FindStringSubmatch(msgs[0].Body)
		_, err = s.Ppl.Accept(ctx, acct, m[1])
		s.must(err, "accept "+a.Email)
		a.P = s.principal(a)
		s.must(s.deviceUp(ctx, a), "device "+a.Email)
		s.ByEmail[a.Email] = a
		if _, ok := s.ByName[a.Name]; !ok {
			s.ByName[a.Name] = a
		}
		s.All = append(s.All, a)
		if a.Availability != "" && a.Availability != "Available" {
			s.must(s.Rol.EditMe(ctx, a.P, roles.ProfileEdit{Availability: &a.Availability}), "availability")
		}
	}

	// the owner's own place in the structure and profile (flow f11 is switched off for the moment)
	s.setFlow(ctx, "f11", 0)
	req, applied, err := s.Ppl.Update(ctx, o.P, o.Person, people0Changes(s))
	s.must(err, "owner placement")
	if !applied {
		panic(fmt.Sprintf("seed: placement waited for approval: %v", req))
	}
	s.setFlow(ctx, "f11", 2)

	var prof struct {
		Status, Cover, Photo, Phone, Bio string
		Skills                           []string
	}
	s.st("profile", &prof)
	cover := prof.Cover
	s.must(s.Rol.EditMe(ctx, o.P, roles.ProfileEdit{Availability: &prof.Status, Cover: &cover, Bio: &prof.Bio, Skills: &prof.Skills}), "owner profile")
}

func people0Changes(s *Seeder) people.Changes {
	m := s.data.org.Mine
	sec, br, un, dp := s.UnitIDs[m.Sectors], s.UnitIDs[m.Branches], s.UnitIDs[m.Units], s.UnitIDs[m.Departments]
	title := "IT Security & Risk Officer"
	return people.Changes{Sector: &sec, Branch: &br, Unit: &un, Department: &dp, Title: &title}
}

func nPerson(first, last string, d dirPerson, a *Actor, unit func(kind, name string) *uuid.UUID) people.NewPerson {
	genders := []string{"Female", "Male"}
	return people.NewPerson{First: first, Last: last, Email: d.Email, Phone: fmt.Sprintf("+2547%08d", 10000000+d.ID*37), DOB: fmt.Sprintf("%d-%02d-%02d", 1975+d.ID%28, 1+d.ID%12, 1+d.ID%27),
		Gender: genders[d.ID%2], Title: d.Title, Sector: unit("sector", d.Sector), Branch: unit("branch", d.Loc), Unit: unit("unit", d.Unit), Department: unit("department", d.Dept), Role: a.Role, Color: d.Color}
}

func nPersonRaw(a *Actor) people.NewPerson {
	return people.NewPerson{First: a.First, Last: a.Last, Email: a.Email, Phone: "+254700000000", DOB: "1980-01-01", Gender: "Other", Title: a.Title, Role: a.Role, Color: a.Color}
}

// deviceUp registers an actor's device in the key directory the way the app does on first sign-in.
func (s *Seeder) deviceUp(ctx context.Context, a *Actor) error {
	d, err := s.deviceFor(a.Email)
	if err != nil {
		return err
	}
	a.Dev = d
	pub := d.Public()
	bind := bindSig(d, pub.Identity, a.Account)
	dev, err := s.Keys.Register(ctx, s.principal(a), "Demo device", base64.StdEncoding.EncodeToString(pub.Identity), base64.StdEncoding.EncodeToString(pub.Signing), base64.StdEncoding.EncodeToString(bind))
	if err != nil {
		return err
	}
	a.DevID = dev.ID
	return nil
}

func bindSig(d *e2ee.Device, identity []byte, acct uuid.UUID) []byte {
	return ed25519.Sign(d.Sign, keys.BindingMessage(identity, acct))
}
