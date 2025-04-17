package main

import (
	"github.com/Brialius/gh-topics/pkg/cli"
)

func init() {
	cli.Init()
}

func main() {
	args := cli.Parse()
	cli.Run(args)
}
