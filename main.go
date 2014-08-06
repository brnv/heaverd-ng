package main

import (
	"heaverd-ng/tracker"
	"heaverd-ng/webserver"
	"log"
	"os"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("[log] [heaverd-ng] ")
	log.SetOutput(os.Stdin)

	go webserver.Start(":8081")
	go tracker.Start(":1444")

	for {
		time.Sleep(100)
	}
}
