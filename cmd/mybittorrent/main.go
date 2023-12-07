package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/codecrafters-io/bittorrent-starter-go/logging"
)

const (
	cmdDecode        = "decode"
	cmdInfo          = "info"
	cmdPeers         = "peers"
	cmdHandshake     = "handshake"
	cmdDownloadPiece = "download_piece"
	cmdDownloadFile  = "download"
	logLevel         = logging.LevelInfo
)

var logger = logging.New(logLevel)

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
	case cmdDownloadFile:
		doDownloadFile()
	default:
		fmt.Printf("Unknown command %q\n", command)
		os.Exit(1)
	}
}

func doDecode() {
	bencodedValue := os.Args[2]
	decoded, err := Decode(bencodedValue)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	printDecodeOutput(decoded)
}

func doInfo() {
	path := os.Args[2]
	tf, err := NewClient(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tf.PrintInfo()
}

func doPeers() {
	path := os.Args[2]
	tf, err := NewClient(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	PrintPeers(tf.Peers)
}

func doHandshake() {
	if len(os.Args) < 4 {
		fmt.Println("Insufficient number of arguments given.")
		os.Exit(1)
	}
	path := os.Args[2]
	c, err := NewClient(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	peer := os.Args[3] // peer ip_address:port

	logger.Info("Connecting to peer at %s...\n", peer)
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	handshake, err := Handshake(conn, c.InfoHash)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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

	c, err := NewClient(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// TODO manage connections to multiple peers
	conn, err := c.Connect(1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer c.Close(conn)

	// Handshake and run preliminary protocol.
	if err := c.initiateDownload(conn); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Download piece.
	logger.Info("Downloading piece %d from %s to %s\n", piece, path, outputPath)
	if err := c.DownloadPiece(conn, piece, outputPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Piece %d downloaded to %s.\n", piece, outputPath)
}

func doDownloadFile() {
	if len(os.Args) < 5 || os.Args[2] != "-o" {
		fmt.Println("Syntax: mybittorrent download -o " +
			"[OUTPUT_PATH] [TORRENT_PATH]")
		os.Exit(1)
	}
	outputPath := os.Args[3]
	path := os.Args[4]

	c, err := NewClient(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// TODO manage connections to multiple peers
	conn, err := c.Connect(1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer c.Close(conn)

	logger.Info("Downloading %s from %s to %s...\n", c.Info.Name, path, outputPath)
	if err := c.DownloadFile(conn, outputPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Downloaded %s to %s.\n", c.Info.Name, outputPath)
}
