package main

import "qbit-cli/cmd"

var (
	Version = "none"
)

func main() {
	cmd.Execute(Version)
}
