package main

import (
	"context"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "example.com/file-walker/internal/pb"
)

const (
	serverAddr      = "10.114.32.49:50051" // Replace with your server's address and port
	destinationPath = "/home/imafish/tmp/console.output"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileWatcherClient(conn)

	stream, err := client.Subscribe(context.Background(), &pb.SubscribeRequest{})
	if err != nil {
		log.Fatalf("could not subscribe to file changes: %v", err)
	}

	// Listen for file change events
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			// Stream closed by the server
			break
		}
		if err != nil {
			log.Fatalf("error receiving file change event: %v", err)
		}

		// Write the content to destination file
		file, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Failed to open file for writing: %s", err.Error())
		} else {
			file.Close()
			file.WriteString(event.Content)
		}
	}
}
