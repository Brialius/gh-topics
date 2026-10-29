package main

import "github.com/Brialius/gh-topics/pkg/cli"

// version is populated by release builds using -ldflags or a generated source file.
var version = "dev"

func main() {
	cli.Version = version
	cli.Init()
	args := cli.Parse()
	cli.Run(args)
}
