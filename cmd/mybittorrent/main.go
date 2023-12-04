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

	// Decode does not use a torrent file as an argument.
	if command == "decode" {
		bencodedValue := os.Args[2]
		decoded, err := Decode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}
		printDecodeOutput(decoded)
		return
	}

	tf, err := NewTorrentFile(os.Args[2])
	if err != nil {
		fmt.Println(err)
	}
	_ = tf

	switch command {
	case "info":
		path := os.Args[2]
		tf, err := Info(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		printInfoOutput(tf)
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
		handshake, err := Handshake(path, peer)
		if err != nil {
			fmt.Println(err)
			return
		}
		PrintHandshakeOutput(handshake)
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
