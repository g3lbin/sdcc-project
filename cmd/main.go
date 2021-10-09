package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type peer struct {
	containerID string
	publicIP    string
	privateIP   string
	port 		uint16
}

func errorHandler(foo string, err error) {
	log.Fatalf("%s has failed: %s", foo, err)
}

func connectionHandler(p peer, ch chan string, delay bool) {
	var min, max int

	if delay {
		min = 0
		max = 30
	}
	service := p.publicIP + ":" + strconv.Itoa(int(p.port))
	conn, err := net.Dial("tcp", service)
	if err != nil {
		errorHandler("Dial", err)
	}
	defer conn.Close()
	for {
		msg := <- ch
		if delay {
			r := rand.Intn(max - min + 1) + min
			time.Sleep(time.Duration(r) * time.Second)
		}
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
		//buf := new(strings.Builder)
		buf := &bytes.Buffer{}
		_, err = stdcopy.StdCopy(buf, nil, out)
		// _, err = io.Copy(buf, out)
		if err != nil {
			panic(err)
		}
		// check errors
		fmt.Println(r.Replace(buf.String()))
		out.Close()
	}
}

func parseCmdLine() (bool, bool) {
	verbose := false
	delay := false

	args := os.Args
	length := len(os.Args)
	if length > 1 {
		for i := 1; i < length; i++ {
			arg := args[i]
			if arg == "-h" || arg == "--h" || arg == "-help" || arg == "--help" {
				fmt.Println("Usage: ./launcher [OPTION]...")
				fmt.Println("\nAvailable options:")
				fmt.Println("\t-v, --verbose\t Print logging information")
				fmt.Println("\t-d, --delay\t Enable an arbitrary delay in sending messages")
				fmt.Println("\t-h, --help\t View this message")
				os.Exit(0)
			} else if arg == "-v" || arg == "--verbose" {
				verbose = true
			} else if arg == "-d" || arg == "--delay" {
				delay = true
			}
		}
	}
	return verbose, delay
}

func main() {
	var peers map[string]peer
	var channelMap map[string]chan string
	var replaceList []string
	var temp peer
	var username string
	var verbose bool
	var delay bool

	verbose, delay = parseCmdLine()
	fmt.Println(verbose)
	if delay {
		rand.Seed(time.Now().UnixNano())
	}

	scanner := bufio.NewScanner(os.Stdin)

	channelMap = make(map[string]chan string)
	peers = make(map[string]peer)
	peerNum := 0

	//cmd := exec.Command("docker-compose", "-f", "deployments/docker-compose.yml", "up", "-d", "--build")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	// cmd.Run()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Wait for initialization...\n")
	for ok := true; ok; {
		contJson, err := cli.ContainerInspect(ctx, "registry")
		if err != nil {
			continue
		}
		ok = contJson.State.Status != "exited"
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Set the usernames for the chat participants")
	for _, container := range containers {
		if container.Image != "peer_service" {
			continue
		}
		for ok := true; ok; {
			fmt.Printf("username #%d\n", peerNum + 1)
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

		replaceList = append(replaceList, container.NetworkSettings.Networks["deployments_my_net"].IPAddress + "]")
		var b strings.Builder
		fmt.Fprintf(&b, "%-17s", username + "]")
		replaceList = append(replaceList, b.String())

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
	r := strings.NewReplacer(replaceList...)

	// establish connections with peers
	for user := range peers {
		go connectionHandler(peers[user], channelMap[user], delay)
	}

	fmt.Println()
	printLogs(cli, ctx, containers, r, "sequencer_service")
	fmt.Println("\nNow you can send messages from each peer you want")
	for {
		fmt.Println("\n************* Allowed Actions *************")
		fmt.Println("1) Send a message from a participant")
		fmt.Println("2) View messages received by a participant")
		fmt.Println("3) Quit")

		fmt.Printf("\nSelect an option\n>> ")
		scanner.Scan()
		opt := scanner.Text()

		if opt == "1" || opt == "2" {
			for ok := false; !ok; {
				fmt.Printf("\nEnter an existing username for a chat participant\n>> ")
				scanner.Scan()
				username = scanner.Text()
				_, ok = peers[username]
			}
			if opt == "1" {
				fmt.Printf("Insert a message\n>> ")
				scanner = bufio.NewScanner(os.Stdin)
				scanner.Scan()
				channelMap[username] <- scanner.Text()
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