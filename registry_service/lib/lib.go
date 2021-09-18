package lib

import (
	"bufio"
	"log"
	"os"
	"sync"
)

type Receiver struct {
	Name string
	Str Message
}
type Registry struct {
	FilePath string
	MembersNum int
}

type Message string
type Result int

var fileLock sync.RWMutex
var counterLock sync.RWMutex
var members = 0

func (registry *Registry) RegisterMember(arg Message, res *[]string) error {
	fileLock.Lock()
	file, err := os.OpenFile(registry.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed creating file: %s", err)
	}

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(string(arg) + "\n")
	if err != nil {
		log.Fatalf("WriteString error: %s", err)
	}
	writer.Flush()
	file.Close()
	fileLock.Unlock()

	counterLock.Lock()
	members++
	counterLock.Unlock()

	for {
		counterLock.RLock()
		if members == registry.MembersNum {
			counterLock.RUnlock()
			fileLock.RLock()
			file, err = os.OpenFile(registry.FilePath, os.O_RDONLY, 0644)

			if err != nil {
				log.Fatalf("Failed creating file: %s", err)
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				*res = append(*res, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				log.Fatal("Scanner error: ", err)
			}
			file.Close()
			fileLock.RUnlock()
			break
		}
		counterLock.RUnlock()
	}
	return nil
}

//func (rcv *Sender) Send_message(arg Message, res *Result) error {
//
//}

func (rcv *Receiver) Receive_message(arg Message, res *Result) error {

	rcv.Str = arg
	log.Printf("%s received: %s", rcv.Name, rcv.Str)
	*res = 777

	return nil
}
