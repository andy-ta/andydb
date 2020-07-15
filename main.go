package main

import (
	"github.com/andy-ta/andydb/app"
)

func main() {

	server := &app.App{}
	server.Initialize()
	server.Run(":42069")
}
