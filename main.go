package main

import (
	"flag"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func publishMQTT(broker, topic, payload string) error {
	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetClientID("planted-pi")
	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	defer client.Disconnect(250)
	token = client.Publish(topic, 0, false, payload)
	token.Wait()
	return token.Error()
}

func main() {
	service := flag.String("service", "", "Service name (ato, fertilizer, etc.)")
	dir := flag.String("dir", "/var/run/planted-pi-services", "Service name (ato, fertilizer, etc.)")
	mqttBroker := flag.String("mqtt", "tcp://localhost:1883", "MQTT broker URL")
	relay := flag.Int("relay", 2, "Relay number (1 or 2)")
	timer := flag.Int("timer", 0, "Timer of seconds to off the relay")
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

	pulseTimeTopic := fmt.Sprintf("cmnd/relays/PulseTime%d", *relay)
	err := publishMQTT(*mqttBroker, pulseTimeTopic, strconv.Itoa(*timer))
	if err != nil {
		fmt.Printf("Error setting PulseTime: %v\n", err)
	} else {
		fmt.Printf("Set PulseTime%d to %d seconds\n", *relay, *timer)
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

	var topic = fmt.Sprintf("cmnd/relays/POWER%d", *relay)
	fmt.Printf("Watching %s for changes...\n", file)

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
				changed := strings.TrimSpace(string(data))
				if changed == "" {
					fmt.Println("File is empty, no changes detected.")
					continue
				}
				val, err := strconv.Atoi(changed)
				if err != nil {
					fmt.Println("Error converting file content to integer:", err)
					return
				}

				fmt.Printf("Service %s: Value changed to %d \n", *service, val)
				payload := "OFF"
				if val >= 1 {
					payload = "ON"
				}
				err = publishMQTT(*mqttBroker, topic, payload)
				if err != nil {
					fmt.Println("Error publishing MQTT message:", err)
				} else {
					fmt.Printf("Sent MQTT command: %s -> %s\n", topic, payload)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Watcher error:", err)
		}
	}

}
