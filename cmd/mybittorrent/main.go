package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Insufficient number of arguments given.")
		os.Exit(1)
	}
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
	case "handshake":
		if len(os.Args) < 4 {
			fmt.Println("Insufficient number of arguments given.")
			os.Exit(1)
		}
		path := os.Args[2]
		peer := os.Args[3]
		fmt.Println(path, peer)
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
