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
		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}
		printDecodeOutput(decoded)

	case "info":
		path := os.Args[2]
		tf, err := parseTorrentFile(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		printInfoOutput(tf)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
