package rpcregistration

import (
	"bufio"
	"os"
	"sync"

	"github.com/sdcc-project/internal/pkg/utils"
)

type Registry struct {
	FilePath string					// file to maintain the membership
	MembersNum int
}

var fileLock sync.RWMutex			// RWMutex to synchronize the operations on the file
var counterLock sync.RWMutex		// RWMutex to synchronize the accesses to members variable
var members = 0						// number of registered members

// waitForAll does an active wait until all members are registered, then it returns the complete list of members
func waitForAll(registry *Registry) []string {
	var list []string

	for {													// start active wait
		counterLock.RLock()
		if members == registry.MembersNum {
			counterLock.RUnlock()
			fileLock.RLock()
			file, err := os.OpenFile(registry.FilePath, os.O_RDONLY, 0644)

			if err != nil {
				utils.ErrorHandler("OpenFile", err)
			}

			// read all lines of the file (each line corresponds to a single member)
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				list = append(list, scanner.Text())
			}
			if err = scanner.Err(); err != nil {
				utils.ErrorHandler("Scanner", err)
			}
			file.Close()
			fileLock.RUnlock()
			break
		}
		counterLock.RUnlock()
	}
	return list
}

// RetrieveMembership returns in res the memberhip when all members are registered
func (registry *Registry) RetrieveMembership(arg string, res *[]string) error {
	if arg == "" {
		*res = waitForAll(registry)
	}

	return nil
}

// RegisterMember adds a new member and returns in res the memberhip when all members are registered
func (registry *Registry) RegisterMember(arg string, res *[]string) error {
	fileLock.Lock()
	file, err := os.OpenFile(registry.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.ErrorHandler("OpenFile", err)
	}

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(string(arg) + "\n")		// add the new member to the membership
	if err != nil {
		utils.ErrorHandler("WriteString", err)
	}
	writer.Flush()
	file.Close()
	fileLock.Unlock()

	counterLock.Lock()
	members++											// increment the number of registered members
	counterLock.Unlock()

	*res = waitForAll(registry)							// put the member in wait for others and give him the membership

	return nil
}