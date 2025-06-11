package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	service := flag.String("service", "", "Service name (ato, fertilizer, etc.)")
	flag.Parse()

	if *service == "" {
		fmt.Println("No service specified")
		os.Exit(1)
	}

	dir := "/var/run/planted-pi-services"
	file := filepath.Join(dir, *service)

	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Println("Error creating dir:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(file, []byte("0"), 0644); err != nil {
		fmt.Println("Error writing file:", err)
		os.Exit(1)
	}

	for {
		fmt.Printf("%s service running on ARMv6...\n", *service)
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}
		fmt.Println("File content:", string(data))

		time.Sleep(5 * time.Second)
	}
}
