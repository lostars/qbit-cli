package main

import (
	"qbit-cli/internal/cmd"
	_ "qbit-cli/internal/job"
)

var (
	Version = "none"
)

func main() {
	cmd.Execute(Version)
}
