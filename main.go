package main

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	service := flag.String("service", "", "Service name (ato, fertilizer, etc.)")
	dir := flag.String("dir", "/var/run/planted-pi-services", "Service name (ato, fertilizer, etc.)")
	flag.Parse()

	if *service == "" {
		fmt.Println("No service specified")
		os.Exit(1)
	}

	file := filepath.Join(*dir, *service)

	if err := os.MkdirAll(*dir, 0755); err != nil {
		fmt.Println("Error creating dir:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(file, []byte("0"), 0644); err != nil {
		fmt.Println("Error writing file:", err)
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		os.Exit(1)
	}
	defer watcher.Close()

	err = watcher.Add(file)
	if err != nil {
		fmt.Println("Error adding file to watcher:", err)
		os.Exit(1)
	}

	fmt.Printf("Watching %s for changes...\n", file)
	lastValue := 0

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				data, err := os.ReadFile(file)
				if err != nil {
					fmt.Println("Error reading file:", err)
					return
				}
				val, err := strconv.Atoi(strings.TrimSpace(string(data)))
				if err != nil {
					fmt.Println("Error converting file content to integer:", err)
					return
				}
				if lastValue != val {
					fmt.Printf("Service %s: Value changed!\n", *service)
				}
				lastValue = val
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Watcher error:", err)
		}
	}

}
