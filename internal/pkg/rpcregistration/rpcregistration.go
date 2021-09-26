package rpcregistration

import (
	"bufio"
	"os"
	"sync"

	"github.com/sdcc-project/internal/pkg/utils"
)

type Registry struct {
	FilePath string
	MembersNum int
}

var fileLock sync.RWMutex
var counterLock sync.RWMutex
var members = 0

func waitForAll(registry *Registry) []string {
	var list []string

	for {
		counterLock.RLock()
		if members == registry.MembersNum {
			counterLock.RUnlock()
			fileLock.RLock()
			file, err := os.OpenFile(registry.FilePath, os.O_RDONLY, 0644)

			if err != nil {
				utils.ErrorHandler("OpenFile", err)
			}

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

func (registry *Registry) RetrieveMembership(arg string, res *[]string) error {
	if arg == "get me the list of members" {
		*res = waitForAll(registry)
	}

	return nil
}

func (registry *Registry) RegisterMember(arg string, res *[]string) error {
	fileLock.Lock()
	file, err := os.OpenFile(registry.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.ErrorHandler("OpenFile", err)
	}

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(string(arg) + "\n")
	if err != nil {
		utils.ErrorHandler("WriteString", err)
	}
	writer.Flush()
	file.Close()
	fileLock.Unlock()

	counterLock.Lock()
	members++
	counterLock.Unlock()

	*res = waitForAll(registry)

	return nil
}