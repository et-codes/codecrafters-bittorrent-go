package main

import "os"

type TorrentInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type TorrentFile struct {
	Announce    string      `bencode:"announce"`
	Info        TorrentInfo `bencode:"info"`
	InfoHash    string
	Peers       []string // List of peer IP addresses
	PieceHashes []string // SHA-1 hashes of Pieces
}

// NewTorrentFile reads a torrent file and populates the a TorrentFile struct.
func NewTorrentFile(path string) (TorrentFile, error) {
	tf := TorrentFile{}

	data, err := os.ReadFile(path)
	if err != nil {
		return tf, err
	}

	err = decodeTorrentFile(string(data), &tf)
	if err != nil {
		return tf, err
	}

	err = tf.HashInfo()
	if err != nil {
		return tf, err
	}

	tf.HashPieces()

	err = tf.peerList()
	if err != nil {
		return tf, err
	}

	return tf, nil
}
