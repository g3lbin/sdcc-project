/*
This program allows to test the functioning of the implemented algorithms in case there is only one participant
which sends the multicast messages and in case multiple participants simultaneously send the multicast messages
 */
package main

import (
	"flag"
	"fmt"
	"os"

	t "github.com/sdcc-project/test/testlib"
)

// isFlagPassed returns "true" if flagName was specified on the command-line, "false" otherwise
func isFlagPassed(flagName string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == flagName {
			found = true
		}
	})
	return found
}

func main() {
	// define the flags with specified name, default value, and usage string
	algorithmPtr := flag.Int("algorithm", 1, "The multicast algorithm can be 1, 2 or 3")
	sourcePtr := flag.String("source", "single", "The source that sends the messages can be \"single\" or \"multiple\"")
	// parse the command-line flags
	flag.Parse()
	// exit if the necessary flags have not been correctly specified
	if !isFlagPassed("algorithm") || !isFlagPassed("source") ||
		*algorithmPtr < 1 || *algorithmPtr > 3 ||
		(*sourcePtr != "single" && *sourcePtr != "multiple") {
		fmt.Printf("Usage of %s\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	// execute the specific test
	if *sourcePtr == "single" {
		t.TestSingleSource()
	} else {
		if *algorithmPtr == 3 {
			t.TestMultipleSourceCausOrdered()
		} else {
			t.TestMultipleSourceTotOrdered()
		}
	}
}
