package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"os"
	"os/exec"
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "notify":
		message := flag.Arg(1)
		cmd := exec.Command("./serf", "query", "notify", message)
		cmd.Run()
	case "receive-message":
		message, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		dialer, err := net.Dial("tcp", ":1444")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(dialer, string(message))
	default:
	}
}
