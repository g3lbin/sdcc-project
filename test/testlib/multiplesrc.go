package testlib

import (
	"bytes"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// TestMultipleSourceTotOrdered tests the system in case multiple participants simultaneously send the multicast
// messages using a totally ordered multicast algorithm
func TestMultipleSourceTotOrdered() {
	var ds DockerStr
	var r *strings.Replacer
	var peers map[string]Peer
	var channelMap map[string]chan string // map to send messages to the connectionHandler of each peer

	peers = make(map[string]Peer)
	channelMap = make(map[string]chan string)

	// list of multicast messages
	messageList  := [10]string{
		"0",
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
	}

	// test initialization
	peersNum := TestInit(&peers, &channelMap, &r, &ds)
	// establish connections with peers
	for user := range peers {
		go ConnectionHandler(peers[user], channelMap[user])
	}

	fmt.Println("Test started!")
	// send the messages from all participants in parallel
	for i, msg := range messageList {
		user := "user" + strconv.Itoa(i % peersNum)
		channelMap[user] <- msg
	}

	fmt.Println("All messages have been sent, wait for propagation...")
	time.Sleep(3*time.Second)

	fmt.Println("\nTest results:")
	// print received messages from each participant and check the results
	results := make([]string, peersNum)
	for i := 0; i < peersNum; i++ {
		user := "user" + strconv.Itoa(i)
		fmt.Println("\n************* "+user+" *************")
		results[i] = PrintLogs(ds.Cli, ds.Ctx, ds.Containers, r, peers[user].ContainerID)
	}
	if isTestPassed(3, "multiple", results) {
		fmt.Println("\nTest PASSED!")
	} else {
		fmt.Println("\nTest FAILED!")
	}
}

// TestMultipleSourceCausOrdered tests the system in case multiple participants simultaneously send the multicast
// messages using a causally ordered multicast algorithm
func TestMultipleSourceCausOrdered() {
	var ds DockerStr
	var r *strings.Replacer
	var peers map[string]Peer
	var channelMap map[string]chan string // map to send messages to the connectionHandler of each peer

	peers = make(map[string]Peer)
	channelMap = make(map[string]chan string)

	// define causally related multicast messages
	catOwnerMsg := [2]string{
		"Oh no! My cat just jumped out the window.",
		"Whew, the catnip plant broke her fall.",
	}
	friendMsg := "I love when that happens to cats!"

	// test initialization
	peersNum := TestInit(&peers, &channelMap, &r, &ds)
	// establish connections with peers
	for user := range peers {
		go ConnectionHandler(peers[user], channelMap[user])
	}

	fmt.Println("Test started!")
	// send the messages from the cat owner
	for _, msg := range catOwnerMsg {
		channelMap["user0"] <- msg
	}
	// wait until the friend receives both messages
	go func() {
		c := 0
		for ok := true; ok; ok = c != 2 {
			out, err := ds.Cli.ContainerLogs(ds.Ctx, peers["user1"].ContainerID, types.ContainerLogsOptions{ShowStdout: true})
			if err != nil {
				panic(err)
			}

			buf := &bytes.Buffer{}
			_, err = stdcopy.StdCopy(buf, nil, out)

			if err != nil {
				panic(err)
			}

			s := r.Replace(buf.String())
			c = strings.Count(s, "user0")
			out.Close()
		}
		// the friend sends the reply
		channelMap["user1"] <- friendMsg
	}()
	time.Sleep(10*time.Millisecond)
	// the others send a message that is not causally related
	rand.Seed(time.Now().UnixNano())
	for i := 2; i < peersNum; i++ {
		user := "user" + strconv.Itoa(i)
		go func() {
			r := rand.Intn(10)
			time.Sleep(time.Duration(r) * time.Millisecond)

			channelMap[user] <- DummyMsg
		}()
	}

	fmt.Println("All messages have been sent, wait for propagation...")
	time.Sleep(5*time.Second)

	fmt.Println("\nTest results:")
	// print received messages from each participant and check the results
	results := make([]string, peersNum)
	for i := 0; i < peersNum; i++ {
		user := "user" + strconv.Itoa(i)
		fmt.Println("\n************* "+user+" *************")
		results[i] = PrintLogs(ds.Cli, ds.Ctx, ds.Containers, r, peers[user].ContainerID)
	}
	if isTestPassed(3, "multiple", results) {
		fmt.Println("\nTest PASSED!")
	} else {
		fmt.Println("\nTest FAILED!")
	}
}