package main

import "os"

type Client struct {
	Announce    string      `bencode:"announce"`
	Info        TorrentInfo `bencode:"info"`
	InfoHash    string
	Peers       []string // List of peer IP addresses
	PieceHashes []string // SHA-1 hashes of Pieces
}

// NewClient reads a torrent file and populates the a Client struct.
func NewClient(path string) (Client, error) {
	c := Client{}

	data, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = decodeTorrentFile(string(data), &c)
	if err != nil {
		return c, err
	}

	err = c.HashInfo()
	if err != nil {
		return c, err
	}

	c.HashPieces()

	err = c.peerList()
	if err != nil {
		return c, err
	}

	return c, nil
}
