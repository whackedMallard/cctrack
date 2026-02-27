package main

import "github.com/ksred/cctrack/cmd"

func main() {
	cmd.WebFSFunc = WebFS
	cmd.Execute()
}
