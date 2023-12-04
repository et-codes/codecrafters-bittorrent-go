package main

import (
	"fmt"
	"os"
)

const (
	peerID       = "00112233445566778899" // Peer ID used for this client (20 bytes)
	cmdDecode    = "decode"
	cmdInfo      = "info"
	cmdPeers     = "peers"
	cmdHandshake = "handshake"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Insufficient number of arguments given.")
		os.Exit(1)
	}
	command := os.Args[1]

	// Decode does not use a torrent file as an argument.
	if command == cmdDecode {
		bencodedValue := os.Args[2]
		decoded, err := Decode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}
		printDecodeOutput(decoded)
		return
	}

	path := os.Args[2]
	tf, err := NewTorrentFile(path)
	if err != nil {
		fmt.Println(err)
	}

	switch command {
	case cmdInfo:
		tf.PrintInfo()
	case cmdPeers:
		tf.PrintPeers()
	case cmdHandshake:
		if len(os.Args) < 4 {
			fmt.Println("Insufficient number of arguments given.")
			os.Exit(1)
		}
		peer := os.Args[3]
		handshake, err := tf.Handshake(peer)
		if err != nil {
			fmt.Println(err)
			return
		}
		PrintHandshake(handshake)
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
