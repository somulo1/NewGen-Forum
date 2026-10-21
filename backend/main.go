// Hub backend. `go build -o app && ./app` (or `go run .`) is the whole deployment: it applies the database migrations, installs
// the reference data, runs the background jobs and serves the API. See internal/runapi for the switches.
package main

import "hub/internal/runapi"

func main() { runapi.Run() }
