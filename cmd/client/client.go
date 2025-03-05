package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"internal/pb"
	"internal/stringutil"
)

var lck sync.Mutex
var wg sync.WaitGroup

func writeFile(destinationPath string, content string, serverName string) {
	lck.Lock()
	defer lck.Unlock()

	file, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open file for writing: %s", err.Error())
	} else {
		content = fmt.Sprintf("[%s] %s", serverName, content)
		log.Printf("Update content: %s", content)
		file.WriteString(content)
		file.Close()
	}
}

func doClient(serverAddr string, serverName string, destinationPath string) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	defer conn.Close()
	log.Printf("connected to %s\n", serverAddr)

	client := pb.NewFileWatcherClient(conn)

	stream, err := client.Subscribe(context.Background(), &pb.SubscribeRequest{})
	if err != nil {
		log.Printf("could not subscribe to file changes: %v", err)
		return
	}

	// Listen for file change events
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			// Stream closed by the server
			log.Printf("connection closed by server..")
			break
		}
		if err != nil {
			log.Printf("error receiving file change event: %v", err)
			break
		}

		strippedString := stringutil.StripColorCodes(event.Content)

		// Write the content to destination file
		writeFile(destinationPath, strippedString, serverName)
	}
}

func oneClient(serverAddr string, serverName string, destinationPath string) {
	defer wg.Done()

	errCnt := 0
	for {
		now := time.Now()

		doClient(serverAddr, serverName, destinationPath)

		errCnt += 1
		elapsed := time.Now().Sub(now).Seconds()
		if elapsed > 900 {
			errCnt = 0
		}
		if errCnt > 10 {
			errCnt = 10
		}
		waitTime := 1 << errCnt
		log.Printf("restarting client in %d seconds ...\n", waitTime)
		time.Sleep(time.Second * time.Duration(waitTime))
	}

}

type StringSlice []string

func (s *StringSlice) String() string {
	return strings.Join(*s, " ")
}

func (s *StringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func main() {
	var destinationPath string
	var servers StringSlice
	var serverNames StringSlice

	flag.Var(&servers, "s", "Specify multiple values (can be used multiple times)")
	flag.Var(&serverNames, "n", "Specify a name for each server (can be used multiple times)")

	flag.StringVar(&destinationPath, "d", filepath.Join(os.Getenv("HOME"), ".tmp", "console.output"), "target file to save the content from server")
	flag.Parse()

	log.Printf("destination file: %s\n", destinationPath)

	if len(servers) == 0 {
		log.Printf("You didn't specify any servers.")
		flag.Usage()
		os.Exit(1)
	}

	namesCount := len(serverNames)
	for i, server := range servers {
		var serverName string
		if i < namesCount {
			serverName = serverNames[i]
		} else {
			serverName = fmt.Sprintf("#%d", i)
			log.Printf("using default server name [%s] for server %s.", serverName, server)
		}
		log.Printf("server %s: %s", serverName, server)

		wg.Add(1)
		go oneClient(server, serverName, destinationPath)
	}

	wg.Wait()

}
