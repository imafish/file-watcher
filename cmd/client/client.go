package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"example.com/file-walker/internal/pb"
	"example.com/file-walker/internal/stringutil"
)

func doClient(serverAddr string, destinationPath string) {
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
		file, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Failed to open file for writing: %s", err.Error())
		} else {
			log.Printf("Update content: %s", strippedString)
			file.WriteString(strippedString)
			file.Close()
		}
	}
}

func main() {
	var serverAddr string
	var destinationPath string
	flag.StringVar(&serverAddr, "s", "10.114.32.49:50051", "file-watcher server address")
	flag.StringVar(&destinationPath, "d", filepath.Join(os.Getenv("HOME"), ".tmp", "console.output"), "target file to save the content from server")
	flag.Parse()

	log.Printf("server address: %s\n", serverAddr)
	log.Printf("destination file: %s\n", destinationPath)

	errCnt := 0
	for {
		now := time.Now()

		doClient(serverAddr, destinationPath)

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
