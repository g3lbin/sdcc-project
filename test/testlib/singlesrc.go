package testlib

import (
"fmt"
"strconv"
"strings"
"time"
)

// TestSingleSource tests the system in case there is only one participant which sends the multicast messages
func TestSingleSource() {
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
	// send the messages from only one participant
	for _, msg := range messageList {
		channelMap["user0"] <- msg
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
