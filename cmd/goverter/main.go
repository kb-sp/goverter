package main

import (
	"os"

	"github.com/kb-sp/goverter/cli"
)

func main() {
	cli.Run(os.Args, cli.RunOpts{})
}
