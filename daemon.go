package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"time"
)

type Message struct {
	Name, Text string
}

func main() {
	l, err := net.Listen("tcp", ":1444")
	if err != nil {
		fmt.Println(err)
		return
	}
	go datasender()

	for {
		socket, _ := l.Accept()
		dec := json.NewDecoder(socket)
		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", m)
		socket.Close()
	}
}

func datasender() {
	const jsonStream = `
		{"Name": "Ed", "Text": "Knock knock."}
	`
	for {
		send(jsonStream)
		time.Sleep(time.Second * 1)
	}
}

func send(message string) {
	cmd := exec.Command("./heaverd-query.sh", "send", string(message))
	out, _ := cmd.Output()

	fmt.Println(string(out))
}
