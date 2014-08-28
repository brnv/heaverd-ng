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
	f, err := os.OpenFile("./query.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal("[error]", err)
	}
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("[heaverd-tracker-query] ")
	log.SetOutput(f)

	var clusterPort string
	flag.StringVar(&clusterPort, "cluster-port", "1444", "")

	var action string
	flag.StringVar(&action, "action", "notify", "notify|receive-message|intent")

	var message string
	flag.StringVar(&message, "message", "", "")

	flag.Parse()

	switch action {
	case "notify":
		cmd := exec.Command("serf", "query", "notify", message)
		cmd.Run()
	case "receive-message":
		message, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Println(err)
		}
		dialer, err := net.Dial("tcp", ":"+clusterPort)
		if err != nil {
			log.Println(err)
		}
		fmt.Fprintf(dialer, string(message))
	case "intent":
		cmd := exec.Command("serf", "query", "intent", message)
		cmd.Run()
	default:
		flag.Usage()
	}
}
