package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	l "github.com/codecrafters-io/bittorrent-starter-go/logger"
)

const (
	cmdDecode        = "decode"
	cmdInfo          = "info"
	cmdPeers         = "peers"
	cmdHandshake     = "handshake"
	cmdDownloadPiece = "download_piece"
	logLevel         = l.LevelDebug
)

var logger = l.New(logLevel)

func main() {
	if len(os.Args) < 3 {
		logger.Fatal("Insufficient number of arguments given.\n")
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
		logger.Fatal("Unknown command %q\n", command)
	}
}

func doDecode() {
	bencodedValue := os.Args[2]
	decoded, err := Decode(bencodedValue)
	if err != nil {
		logger.Fatal(err.Error())
	}
	printDecodeOutput(decoded)
}

func doInfo() {
	path := os.Args[2]
	tf, err := NewClient(path)
	if err != nil {
		logger.Fatal(err.Error())
	}
	tf.PrintInfo()
}

func doPeers() {
	path := os.Args[2]
	tf, err := NewClient(path)
	if err != nil {
		logger.Fatal(err.Error())
	}
	PrintPeers(tf.Peers)
}

func doHandshake() {
	if len(os.Args) < 4 {
		logger.Fatal("Insufficient number of arguments given.")
	}
	path := os.Args[2]
	c, err := NewClient(path)
	if err != nil {
		logger.Fatal(err.Error())
	}
	peer := os.Args[3] // peer ip_address:port

	logger.Info("Connecting to peer at %s...\n", peer)
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer conn.Close()

	handshake, err := Handshake(conn, c.InfoHash)
	if err != nil {
		logger.Fatal(err.Error())
	}

	PrintHandshake(handshake)
}

func doDownloadPiece() {
	if len(os.Args) < 6 || os.Args[2] != "-o" {
		logger.Fatal("Syntax: mybittorrent download_piece -o " +
			"[OUTPUT_PATH] [TORRENT_PATH] [PIECE_INDEX]")
	}
	outputPath := os.Args[3]
	path := os.Args[4]
	piece, _ := strconv.Atoi(os.Args[5])

	c, err := NewClient(path)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// TODO manage connections to multiple peers
	conn, err := c.Connect(1)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer c.Close(conn)

	logger.Info("Downloading piece %d from %s to %s\n", piece, path, outputPath)
	if err := c.DownloadPiece(conn, piece, outputPath); err != nil {
		logger.Fatal(err.Error())
	}

	fmt.Printf("Piece %d downloaded to %s.\n", piece, outputPath)
}
