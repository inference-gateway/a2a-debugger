package main

import (
	"github.com/inference-gateway/a2a-debugger/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.Execute(version, commit, date)
}
