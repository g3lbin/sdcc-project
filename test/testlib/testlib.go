package testlib

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
	"os/exec"
	"strconv"
	"strings"
)

type Peer struct {
	ContainerID string
	PublicIP    string
	PrivateIP   string
	Port        uint16
}

type DockerStr struct {
	Cli *client.Client
	Ctx context.Context
	Containers []types.Container
}

const DummyMsg = "Dummy message!"

func ErrorHandler(foo string, err error) {
	log.Fatalf("%s has failed: %s", foo, err)
}

// PrintLogs prints the log of the container identified by contID
func PrintLogs(
	cli *client.Client,
	ctx context.Context,
	containers []types.Container,
	r *strings.Replacer,
	contID string,
) string {
	var s string

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

		s = buf.String()
		fmt.Println(r.Replace(s)) // replace the peer's IP address with its username
		out.Close()
	}

	return s
}

// ConnectionHandler establishes a connection with the specified peer p to send the test messages
func ConnectionHandler(p Peer, ch chan string) {
	service := p.PublicIP + ":" + strconv.Itoa(int(p.Port))
	conn, err := net.Dial("tcp", service)
	if err != nil {
		ErrorHandler("Dial", err)
	}
	defer conn.Close()
	for {
		msg := <-ch

		_, err = conn.Write([]byte(msg + "\n"))
		if err != nil {
			ErrorHandler("Write", err)
		}
	}
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// TestInit waits for the system to be up and running to assign the usernames to the peers and to set up some structures
func TestInit(
	peers *map[string]Peer,
	channelMap *map[string]chan string,
	r **strings.Replacer,
	dsAddr *DockerStr,
) int {
	var replaceList []string
	var peersNum = 0 // index of peers (starts from 0)
	var temp Peer
	var err error

	// initialize a new API client
	(*dsAddr).Ctx = context.Background()
	(*dsAddr).Cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		ErrorHandler("NewClientWithOpts", err)
	}
	// wait until the "registry" container has exited (so the peers' registration has completed)
	fmt.Printf("Wait for initialization...\n")
	// active wait
	for ok := true; ok; {
		contJson, err := (*dsAddr).Cli.ContainerInspect((*dsAddr).Ctx, "registry")
		if err != nil {
			continue
		}
		ok = contJson.State.Status != "exited"
	}
	clearScreen()
	// assign a username to each peer
	(*dsAddr).Containers, err = (*dsAddr).Cli.ContainerList((*dsAddr).Ctx, types.ContainerListOptions{})
	for _, container := range (*dsAddr).Containers {
		if container.Image != "peer_service" {
			continue
		}
		replaceList = append(replaceList, container.NetworkSettings.Networks["deployments_my_net"].IPAddress+"]")
		var b strings.Builder
		username := "user"+strconv.Itoa(peersNum)
		fmt.Fprintf(&b, "%-6s", username+"]")
		replaceList = append(replaceList, b.String())
		// assign the username to the container
		(*dsAddr).Cli.ContainerRename((*dsAddr).Ctx, container.ID, username)
		temp.ContainerID = container.ID
		for _, port := range container.Ports {
			if port.PrivatePort == 8888 {
				temp.PublicIP = port.IP
				temp.Port = port.PublicPort
				break
			}
		}
		(*peers)[username] = temp
		peersNum++

		(*channelMap)[username] = make(chan string, 100)
	}
	// create the replacer to print peers' usernames instead of IP addresses
	*r = strings.NewReplacer(replaceList...)

	return peersNum
}

// isTestPassed returns "true" if the test result is as expected, "false" otherwise
func isTestPassed(algo int, source string, results []string) bool {
	var passed = true
	if algo == 1 || (algo == 2 && source == "multiple") {
		for i := 1; i < len(results); i++ {
			// in the totally ordered multicast the messages received must be the same for everyone
			if results[i-1] != results[i] {
				passed = false
				break
			}
		}
	} else if algo == 2 && source == "single" {
		// in the decentralized totally ordered multicast if there isn't at least one message from all
		// the message at the top of the queue cannot be delivered to the application.
		// In this case no message was delivered
		for i := 0; i < len(results); i++ {
			// in this test only "user0" sends messages
			receivedMsgNum := strings.Count(results[i], "[user0]")
			if receivedMsgNum != 0 {
				passed = false
				break
			}
		}
	} else if algo == 3 && source == "single" {
		// the causally ordered multicast provides the messages received must be the same of "user0", because
		// in this case it is the only sender (therefore the messages received by "user0" define the causal relationship)
		for i := 1; i < len(results); i++ {
			if results[i] != results[0] {
				passed = false
				break
			}
		}
	} else {
		// messages that define a causal relationship
		var catOwnerMsg = [2]string{
			"Oh no! My cat just jumped out the window.",
			"Whew, the catnip plant broke her fall.",
		}
		var friendMsg = "I love when that happens to cats!"
		var received = [3]bool{false, false, false}

		for i := 0; i < len(results); i++ {
			scanner := bufio.NewScanner(strings.NewReader(results[i]))
			// check if the causally related messages have been received in the same order
			for scanner.Scan() {
				s := scanner.Text()
				if strings.Contains(s, catOwnerMsg[0]) {
					received[0] = true
				} else if strings.Contains(s, catOwnerMsg[1]) {
					if !received[0] { // the second cat owner's message precedes the first
						passed = false
						break
					}
					received[1] = true
				} else if strings.Contains(s, friendMsg) {
					if !received[0] || !received[1] { // the friend reply message precedes one or both cat owner's messages
						passed = false
						break
					}
					received[2] = true
				}
			}
			if !received[0] || !received[1] || !received[2] { // not all messages have been received
				passed = false
				break
			} else {
				// reset the values for the next results
				received = [3]bool{false, false, false}
			}
		}
	}

	return passed
}