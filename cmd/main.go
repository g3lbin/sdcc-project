package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type peer struct {
	containerID string
	publicIP    string
	privateIP   string
	port        uint16
}

func errorHandler(foo string, err error) {
	log.Fatalf("%s has failed: %s", foo, err)
}

func connectionHandler(p peer, ch chan string) {
	service := p.publicIP + ":" + strconv.Itoa(int(p.port))
	conn, err := net.Dial("tcp", service)
	if err != nil {
		errorHandler("Dial", err)
	}
	defer conn.Close()
	for {
		msg := <-ch

		_, err = conn.Write([]byte(msg + "\n"))
		if err != nil {
			errorHandler("Write", err)
		}
	}
}

func printLogs(cli *client.Client, ctx context.Context, containers []types.Container, r *strings.Replacer, contID string) {
	for _, container := range containers {
		if container.ID != contID {
			continue
		}
		out, err := cli.ContainerLogs(ctx, contID, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			panic(err)
		}

		buf := &bytes.Buffer{}
		_, err = stdcopy.StdCopy(buf, nil, out)

		if err != nil {
			panic(err)
		}
		// check errors
		fmt.Println(r.Replace(buf.String()))
		out.Close()
	}
}

func parseCmdLine() bool {
	verbose := false

	args := os.Args
	length := len(os.Args)
	if length > 1 {
		for i := 1; i < length; i++ {
			arg := args[i]
			if arg == "-h" || arg == "--h" || arg == "-help" || arg == "--help" {
				fmt.Println("Usage: ./launcher [OPTION]...")
				fmt.Println("\nAvailable options:")
				fmt.Println("\t-v, --verbose\t Print logging information")
				fmt.Println("\t-h, --help\t View this message")
				os.Exit(0)
			} else if arg == "-v" || arg == "-V" || arg == "--verbose" {
				verbose = true
			}
		}
	}
	return verbose
}

func verboseLoop(
	scanner *bufio.Scanner,
	peers map[string]peer,
	channelMap map[string]chan string,
	cli *client.Client,
	ctx context.Context,
	containers []types.Container,
	r *strings.Replacer,
) {
	var username string

	for {
		fmt.Println("\n************* Allowed Actions *************")
		fmt.Println("1) Send a message from a participant")
		fmt.Println("2) View messages received from a participant")
		fmt.Println("3) Quit")

		fmt.Printf("\nSelect an option\n>> ")
		scanner.Scan()
		opt := scanner.Text()

		if opt == "1" || opt == "2" {
			for ok := false; !ok; {
				fmt.Printf("\nEnter an existing username for a communication participant\n>> ")
				scanner.Scan()
				username = scanner.Text()
				_, ok = peers[username]
			}
			if opt == "1" {
				fmt.Printf("Insert a message\n>> ")
				scanner = bufio.NewScanner(os.Stdin)
				scanner.Scan()
				channelMap[username] <- scanner.Text()
				fmt.Printf("\nMessage succesfully sent!\n")
			} else {
				printLogs(cli, ctx, containers, r, peers[username].containerID)
			}
		} else if opt == "3" {
			fmt.Println("\nGood bye!")
			os.Exit(0)
		} else {
			fmt.Println("Invalid option!")
			continue
		}
		fmt.Println()
	}
}

func simpleLoop(scanner *bufio.Scanner, peers map[string]peer, channelMap map[string]chan string) {
	var username string

	for {
		for ok := false; !ok; {
			fmt.Printf("\nEnter an existing username for a chat participant\n>> ")
			scanner.Scan()
			username = scanner.Text()
			_, ok = peers[username]
		}
		fmt.Printf("Insert a message\n>> ")
		scanner = bufio.NewScanner(os.Stdin)
		scanner.Scan()
		channelMap[username] <- scanner.Text()
		fmt.Printf("\nMessage succesfully sent!\n\n")
	}
}

func main() {
	var peers map[string]peer
	var channelMap map[string]chan string // map to send messages to the connectionHandler of each peer
	var replaceList []string // list to replace peers' IP addresses with usernames
	var temp peer
	var username string
	var verbose bool
	var peerNum = 0 // index of peers (starts from 0)

	verbose = parseCmdLine()

	channelMap = make(map[string]chan string)
	peers = make(map[string]peer)

	// create the scanner to read lines from stdin
	scanner := bufio.NewScanner(os.Stdin)

	//cmd := exec.Command("docker-compose", "-f", "deployments/docker-compose.yml", "up", "-d", "--build")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	// cmd.Run()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	// wait until the "registry" container has exited (so the peers' registration has completed)
	fmt.Printf("Wait for initialization...\n")
	// active wait
	for ok := true; ok; {
		contJson, err := cli.ContainerInspect(ctx, "registry")
		if err != nil {
			continue
		}
		ok = contJson.State.Status != "exited"
	}
	// assign a username to each peer
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nSet the usernames for the communication participants\n")
	for _, container := range containers {
		if container.Image != "peer_service" {
			continue
		}
		for ok := true; ok; {
			fmt.Printf("username #%d\n", peerNum+1)
			fmt.Printf(">> ")
			scanner.Scan()
			username = scanner.Text()

			for strings.Contains(username, " ") || username == "" {
				fmt.Println("Insert an non-empty username which doesn't contain whitespaces")
				fmt.Printf(">> ")
				scanner.Scan()
				username = scanner.Text()
			}
			if len(username) > 16 {
				fmt.Println("The size of the username cannot exceed 16 characters!")
				continue
			}
			ok = false
			l := len(replaceList)
			if l == 0 {
				break
			}
			for i := 0; i < l/2; i++ {
				if replaceList[2*i+1] == username {
					ok = true
					fmt.Println(("\nUsername already in use!"))
					break
				}
			}
		}
		// replaceList alternately contains the IP addresses and usernames of the peers
		replaceList = append(replaceList, container.NetworkSettings.Networks["deployments_my_net"].IPAddress+"]")
		var b strings.Builder
		fmt.Fprintf(&b, "%-17s", username+"]")
		replaceList = append(replaceList, b.String())
		// assign the username to the container
		cli.ContainerRename(ctx, container.ID, username)
		temp.containerID = container.ID
		for _, port := range container.Ports {
			if port.PrivatePort == 8888 {
				temp.publicIP = port.IP
				temp.port = port.PublicPort
				break
			}
		}
		peers[username] = temp
		peerNum++

		channelMap[username] = make(chan string, 100)
	}
	// create the replacer to print peers' usernames instead of IP addresses
	r := strings.NewReplacer(replaceList...)

	// establish connections with peers
	for user := range peers {
		go connectionHandler(peers[user], channelMap[user])
	}

	fmt.Println()
	printLogs(cli, ctx, containers, r, "sequencer_service")
	fmt.Println("\nNow you can send messages from each peer you want")

	if verbose {
		verboseLoop(scanner, peers, channelMap, cli, ctx, containers, r)
	} else {
		simpleLoop(scanner, peers, channelMap)
	}
}
