package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type peer struct {
	username string
	containerID string
	ipAddr string
	port uint16
}

func errorHandler(foo string, err error) {
	log.Fatalf("%s has failed: %s", foo, err)
}

func getPort(peers []peer, user string, num int) uint16 {
	for i := 0; i < num; i++ {
		if peers[i].username == user {
			return peers[i].port
		}
	}
	return 1
}

func getAddress(peers []peer, user string, num int) string {
	for i := 0; i < num; i++ {
		if peers[i].username == user {
			return peers[i].ipAddr
		}
	}
	return "undefined"
}

func main() {
	var peers []peer
	var temp peer
	var username string
	var service string
	var msg string

	peerNum := 0

	cmd := exec.Command("docker-compose", "-f", "deployments/docker-compose.yml", "up", "-d", "--build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// cmd.Run()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Set the usernames for the chat participants")
	for _, container := range containers {
		if container.Image != "service_client" {
			continue
		}

		fmt.Printf("username #%d\n", peerNum + 1)
		fmt.Printf(">> ")
		fmt.Scanln(&username)
		for strings.Contains(username, " ") {
			fmt.Println("Insert an username which doesn't contain whitespace")
			fmt.Printf(">> ")
			fmt.Scanln(&username)
		}
		temp.username = username
		temp.containerID = container.ID
		for _, port := range container.Ports {
			if port.PrivatePort == 8888 {
				temp.ipAddr = port.IP
				temp.port = port.PublicPort
				break
			}
		}
		peers = append(peers, temp)
		peerNum++
	}

	fmt.Println("\nNow you can send messages from each peer you want")
	for {
		for ok := true; ok; ok = getPort(peers, username, peerNum) == 1 {
			fmt.Printf("\n\nInsert an existent peer's username\n>> ")
			fmt.Scanln(&username)
		}
		port := getPort(peers, username, peerNum)
		addr := getAddress(peers, username, peerNum)
		service = addr + ":" + strconv.Itoa(int(port))
		conn, err := net.Dial("tcp", service)
		if err != nil {
			errorHandler("Dial", err)
		}

		fmt.Printf("Insert a message\n>> ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		msg = scanner.Text()

		_, err = conn.Write([]byte(msg))
		if err != nil {
			errorHandler("Write", err)
		}

		err = conn.Close()
		if err != nil {
			errorHandler("Conn", err)
		}
	}
}