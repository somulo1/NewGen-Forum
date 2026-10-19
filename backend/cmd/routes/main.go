// Command routes prints every HTTP route (used to build docs/API.md).
package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/mail"
	"hub/internal/modules"
	"hub/internal/modules/auth"
	"hub/internal/server"
)

func main() {
	cfg, _ := config.Load()
	dev := &mail.Dev{}
	a, err := app.New(context.Background(), cfg, dev)
	if err != nil {
		panic(err)
	}
	defer a.Close()
	au := auth.New(a)
	h := server.Router(server.Deps{App: a, Auth: au, Dev: dev, Modules: modules.All(a, au)})
	var out []string
	_ = chi.Walk(h.(chi.Routes), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		out = append(out, route+" "+method)
		return nil
	})
	sort.Strings(out)
	for _, o := range out {
		fmt.Println(o)
	}
}
