package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

const (
	cmdDecode        = "decode"
	cmdInfo          = "info"
	cmdPeers         = "peers"
	cmdHandshake     = "handshake"
	cmdDownloadPiece = "download_piece"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Insufficient number of arguments given.")
		os.Exit(1)
	}
	command := os.Args[1]

	switch command {
	case cmdDecode:
		doDecode()
	case cmdInfo:
		doInfo()
	case cmdPeers:
		doPeers()
	case cmdHandshake:
		doHandshake()
	case cmdDownloadPiece:
		doDownloadPiece()
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

func doDecode() {
	bencodedValue := os.Args[2]
	decoded, err := Decode(bencodedValue)
	if err != nil {
		fmt.Println(err)
		return
	}
	printDecodeOutput(decoded)
}

func doInfo() {
	path := os.Args[2]
	tf, err := NewTorrentFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	tf.PrintInfo()
}

func doPeers() {
	path := os.Args[2]
	tf, err := NewTorrentFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	tf.PrintPeers()
}

func doHandshake() {
	if len(os.Args) < 4 {
		fmt.Println("Insufficient number of arguments given.")
		os.Exit(1)
	}
	path := os.Args[2]
	tf, err := NewTorrentFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	peer := os.Args[3]
	// Establish a connection with peer.
	conn, err := net.Dial("tcp", tf.Peers[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	handshake, err := tf.Handshake(conn, peer)
	if err != nil {
		fmt.Println(err)
		return
	}
	PrintHandshake(handshake)
}

func doDownloadPiece() {
	if len(os.Args) < 6 || os.Args[2] != "-o" {
		fmt.Println("Syntax: mybittorrent download_piece -o " +
			"[OUTPUT_PATH] [TORRENT_PATH] [PIECE_INDEX]")
		os.Exit(1)
	}
	outputPath := os.Args[3]
	path := os.Args[4]
	piece, _ := strconv.Atoi(os.Args[5])

	tf, err := NewTorrentFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Downloading piece %d from %s to %s\n", piece, path, outputPath)
	tf.DownloadPiece()
}
