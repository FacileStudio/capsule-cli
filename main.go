package main

import "github.com/FacileStudio/capsule-cli/cmd"

var version = "dev"

func main() {
	cmd.Version = version
	cmd.Execute()
}
