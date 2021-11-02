package main

import (
	"fmt"
	t "github.com/sdcc-project/test/testlib"
	"strconv"
	"strings"
	"time"
)

func main() {
	var ds t.DockerStr
	var r *strings.Replacer
	var peers map[string]t.Peer
	var channelMap map[string]chan string // map to send messages to the connectionHandler of each peer

	peers = make(map[string]t.Peer)
	channelMap = make(map[string]chan string)

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

	peersNum := t.TestInit(&peers, &channelMap, &r, &ds)

	// establish connections with peers
	for user := range peers {
		go t.ConnectionHandler(peers[user], channelMap[user])
	}

	fmt.Println("Test started!")
	// send the messages from all participants in parallel
	for i, msg := range messageList {
		user := "user" + strconv.Itoa(i % peersNum)
		channelMap[user] <- msg
	}

	fmt.Println("All messages have been sent, wait for propagation...")
	time.Sleep(10*time.Second)

	fmt.Println("\nTest results:")
	// print received messages from each participant
	for i := 0; i < peersNum; i++ {
		user := "user" + strconv.Itoa(i)
		fmt.Println("\n************* "+user+" *************")
		t.PrintLogs(ds.Cli, ds.Ctx, ds.Containers, r, peers[user].ContainerID)
	}
}
