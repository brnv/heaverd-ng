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

	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	log.SetPrefix("[heaverd-tracker-query] ")
	log.SetOutput(f)

	var clusterPort string
	flag.StringVar(&clusterPort, "cluster-port", "1444", "")
	var action string
	flag.StringVar(&action, "action", "notify", "notify|receive-message")
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
			fmt.Printf("0")
			return
		}
		dialer, err := net.Dial("tcp", ":"+clusterPort)
		if err != nil {
			fmt.Printf("0")
			return
		} else {
			fmt.Fprintf(dialer, string(message))
			fmt.Printf("1")
			return
		}
	default:
		flag.Usage()
	}

}
