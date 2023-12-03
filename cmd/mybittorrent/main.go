package main

import (
	"fmt"
	"os"
)

func main() {
	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]
		err := Decode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}
	case "info":
		path := os.Args[2]
		err := Info(path)
		if err != nil {
			fmt.Println(err)
			return
		}
	case "peers":
		path := os.Args[2]
		err := Peers(path)
		if err != nil {
			fmt.Println(err)
			return
		}
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
