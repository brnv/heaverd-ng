package main

import (
	"encoding/json"
	"fmt"
	"heaverd-ng/libscore"
	"log"
	"net"
	"os"
	"os/exec"
)

func main() {
	cluster := make(map[string]libscore.Host)
	host := libscore.Host{}
	hostChan := make(chan libscore.Host)

	connection, err := net.Listen("tcp", ":1444")
	if err != nil {
		fmt.Println(err)
		return
	}
	go clusterListener(connection, hostChan)

	for {
		select {
		case hostInfo := <-hostChan:
			cluster[hostInfo.Hostname] = hostInfo
		default:
			err := host.Refresh()
			if err != nil {
				fmt.Printf("Error refreshing my host state: %v\n", err)
				os.Exit(1)
			}
			json, _ := json.Marshal(host)

			cmd := exec.Command("./heaverd-query.sh", "hostinfo", string(json))
			cmd.Output()

			fmt.Println(cluster)
		}
	}

}

func clusterListener(connection net.Listener, hostChan chan libscore.Host) {
	for {
		socket, _ := connection.Accept()
		go parseSocketMessage(socket, hostChan)
	}
}

func parseSocketMessage(socket net.Conn, hostChan chan libscore.Host) {
	dec := json.NewDecoder(socket)
	var host libscore.Host
	err := dec.Decode(&host)
	if err != nil {
		log.Fatal(err)
	}
	hostChan <- host
	socket.Close()
}
