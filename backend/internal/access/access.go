// Package access is the catalog of modules, rights and default roles every new organization starts with.
// It is the server's own copy: the browser may hide a button, but only the server decides.
package access

import (
	_ "embed"
	"encoding/json"
	"sync"
)

//go:embed catalog.json
var raw []byte

type Module struct {
	Key    string   `json:"key"`
	Label  string   `json:"label"`
	Tint   string   `json:"tint"`
	Rights []string `json:"rights"`
}

type Role struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Locked bool   `json:"locked"`
	Needs  string `json:"needs"`
	Scope  string `json:"scope"`
}

type Flow struct {
	ID        string   `json:"id"`
	Mod       string   `json:"mod"`
	Action    string   `json:"action"`
	Levels    int      `json:"levels"`
	Approvers []string `json:"approvers"`
}

type Catalog struct {
	Modules  []Module                       `json:"modules"`
	Roles    []Role                         `json:"roles"`
	Defaults map[string]map[string][]string `json:"defaults"` // role -> module -> allowed rights
	Flows    []Flow                         `json:"flows"`
}

var (
	once sync.Once
	cat  Catalog
)

func Default() *Catalog {
	once.Do(func() {
		if err := json.Unmarshal(raw, &cat); err != nil {
			panic("access: bad catalog: " + err.Error())
		}
	})
	return &cat
}

// Valid reports whether module/right exists in the catalog.
func (c *Catalog) Valid(module, right string) bool {
	for _, m := range c.Modules {
		if m.Key == module {
			for _, r := range m.Rights {
				if r == right {
					return true
				}
			}
		}
	}
	return false
}
