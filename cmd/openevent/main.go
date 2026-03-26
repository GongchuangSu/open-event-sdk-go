package main

import (
	"os"

	"github.com/GongchuangSu/open-event-sdk-go/cmd/openevent/cmd"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	os.Exit(cmd.Execute(version, commit, date))
}
