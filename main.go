package main

import "qbit-cli/cmd"
import _ "qbit-cli/internal/job"

var (
	Version = "none"
)

func main() {
	cmd.Execute(Version)
}
