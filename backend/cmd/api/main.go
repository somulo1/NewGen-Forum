// Command api is the Hub HTTP + realtime API node. Starting it applies migrations, installs reference data and runs the
// background jobs (see internal/runapi), so it is the only thing a small deployment runs.
package main

import "hub/internal/runapi"

func main() { runapi.Run() }
