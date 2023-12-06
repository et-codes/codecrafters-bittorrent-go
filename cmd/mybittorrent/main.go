package main

import (
	"log"
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
		log.Fatal("Insufficient number of arguments given.")
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
		log.Fatal("Unknown command: " + command)
	}
}

func doDecode() {
	bencodedValue := os.Args[2]
	decoded, err := Decode(bencodedValue)
	if err != nil {
		log.Fatal(err)
	}
	printDecodeOutput(decoded)
}

func doInfo() {
	path := os.Args[2]
	tf, err := NewClient(path)
	if err != nil {
		log.Fatal(err)
	}
	tf.PrintInfo()
}

func doPeers() {
	path := os.Args[2]
	tf, err := NewClient(path)
	if err != nil {
		log.Fatal(err)
	}
	PrintPeers(tf.Peers)
}

func doHandshake() {
	if len(os.Args) < 4 {
		log.Fatal("Insufficient number of arguments given.")
	}
	path := os.Args[2]
	c, err := NewClient(path)
	if err != nil {
		log.Fatal(err)
	}
	peer := os.Args[3] // peer ip_address:port

	log.Printf("Connecting to peer at %s...\n", peer)
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	handshake, err := Handshake(conn, c.InfoHash)
	if err != nil {
		log.Fatal(err)
	}

	PrintHandshake(handshake)
}

func doDownloadPiece() {
	if len(os.Args) < 6 || os.Args[2] != "-o" {
		log.Fatal("Syntax: mybittorrent download_piece -o " +
			"[OUTPUT_PATH] [TORRENT_PATH] [PIECE_INDEX]")
	}
	outputPath := os.Args[3]
	path := os.Args[4]
	piece, _ := strconv.Atoi(os.Args[5])

	c, err := NewClient(path)
	if err != nil {
		log.Fatal(err)
	}

	// TODO manage connections to multiple peers
	conn, err := c.Connect(1)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close(conn)

	log.Printf("Downloading piece %d from %s to %s\n", piece, path, outputPath)
	if err := c.DownloadPiece(conn, piece, outputPath); err != nil {
		log.Fatal(err)
	}
}
