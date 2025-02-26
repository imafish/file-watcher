package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"internal/pb"
)

type fileWatcherServer struct {
	pb.UnimplementedFileWatcherServer
	subscribers sync.Map
}

type subscriber struct {
	id      int64
	updates chan string
}

func newServer() *fileWatcherServer {
	return &fileWatcherServer{}
}

func (s *fileWatcherServer) Subscribe(req *pb.SubscribeRequest, stream pb.FileWatcher_SubscribeServer) error {
	sub := &subscriber{
		id:      time.Now().UnixNano(),
		updates: make(chan string, 1),
	}

	s.subscribers.Store(sub.id, sub)
	defer s.subscribers.Delete(sub.id)

	for {
		select {
		case update := <-sub.updates:
			if err := stream.Send(&pb.FileChangeNotification{Content: update}); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func monitorFile(filePath string, timeToWait int, linesToFetch int, notify func(string)) {
	var lastLines []string

	for {
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Error opening file: %v", err)
			time.Sleep(time.Second * time.Duration(timeToWait))
			continue
		}

		stat, err := file.Stat()
		if err != nil {
			log.Printf("Error getting file stats: %v", err)
			file.Close()
			time.Sleep(time.Second * time.Duration(timeToWait))
			continue
		}

		if stat.Size() == 0 {
			file.Close()
			time.Sleep(time.Second * time.Duration(timeToWait))
			continue
		}

		lines, err := readLastNLines(file, linesToFetch)
		file.Close()
		if err != nil {
			log.Printf("Error reading last %d lines: %v", linesToFetch, err)
			time.Sleep(time.Second * time.Duration(timeToWait))
			continue
		}

		if !equalSlices(lastLines, lines) {
			lastLines = lines
			notify(strings.Join(lines, "\n"))
		}

		time.Sleep(time.Second * time.Duration(timeToWait))
	}
}

// readLastNLines reads the last N lines from a file.
func readLastNLines(file *os.File, n int) ([]string, error) {
	if n <= 0 {
		return nil, fmt.Errorf("number of lines to read must be positive")
	}

	// Seek to the end of the file
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	filesize := stat.Size()
	var pos int64 = filesize - 1
	var lines []string
	var line []byte

	for pos >= 0 {
		_, err = file.Seek(pos, os.SEEK_SET)
		if err != nil {
			return nil, err
		}

		char := make([]byte, 1)
		_, err = file.Read(char)
		if err != nil {
			return nil, err
		}

		// Prepend the character
		if char[0] == '\n' {
			if len(line) > 0 {
				// Reverse the line because we read it from end to start
				for i, j := 0, len(line)-1; i < j; i, j = i+1, j-1 {
					line[i], line[j] = line[j], line[i]
				}
				lines = append(lines, string(line))
				line = []byte{}
				if len(lines) == n {
					break
				}
			}
		} else {
			line = append(line, char[0])
		}

		pos--
	}

	// Check if we need to add the last line
	if len(line) > 0 {
		for i, j := 0, len(line)-1; i < j; i, j = i+1, j-1 {
			line[i], line[j] = line[j], line[i]
		}
		lines = append(lines, string(line))
	}

	/*
		// Revert the lines because we read from back to front
		if len(lines) > 1 {
			for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
				lines[i], lines[j] = lines[j], lines[i]
			}
		}
	*/

	return lines, nil
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func main() {
	// Define the command-line flags with default values
	linesToFetch := flag.Int("l", 8, "lines to keep")
	timeToWait := flag.Int("i", 15, "intervals in seconds between each scan")
	port := flag.Int("p", 50051, "port to listen")
	fileToWatch := flag.String("f", os.Getenv("HOME")+"/.tmp/console.output", "file to watch")

	// Parse the flags
	flag.Parse()

	// Print the parsed values
	fmt.Printf("Lines to keep: %d\n", *linesToFetch)
	fmt.Printf("Interval: %d seconds\n", *timeToWait)
	fmt.Printf("Port: %d\n", *port)
	fmt.Printf("File to watch: %s\n", *fileToWatch)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	fileWatcherServer := newServer()
	pb.RegisterFileWatcherServer(s, fileWatcherServer)
	reflection.Register(s)

	go monitorFile(*fileToWatch, *timeToWait, *linesToFetch, func(update string) {
		// Notify all subscribers
		fileWatcherServer.subscribers.Range(func(key, value interface{}) bool {
			sub := value.(*subscriber)
			select {
			case sub.updates <- update:
			default:
			}
			return true
		})
	})

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
