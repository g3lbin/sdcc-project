package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		fmt.Println(image.ID)
	}

	//log.Printf("Inserisci address:port")
	//service := os.Args[1]
	//tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	//checkError(err)
	//conn, err := net.DialTCP("tcp", nil, tcpAddr)
	//checkError(err)
	//_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	//checkError(err)
	//result, err := ioutil.ReadAll(conn)
	//checkError(err)
	//fmt.Println(string(result))
	//os.Exit(0)
}
