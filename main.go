package main

import "github.com/kylecalbert/cctrack/cmd"

func main() {
	cmd.WebFSFunc = WebFS
	cmd.Execute()
}
