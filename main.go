package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("MyApp running on ARMv6...")
		time.Sleep(5 * time.Second)
	}
}
