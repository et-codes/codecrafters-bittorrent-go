package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

type TorrentInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type TorrentFile struct {
	Announce string      `bencode:"announce"`
	Info     TorrentInfo `bencode:"info"`
}

func Info(path string) (TorrentFile, error) {
	tf, err := NewTorrentFile(path)
	if err != nil {
		return tf, err
	}
	return tf, nil
}

// NewTorrentFile populates a TorrentFile struct with info from the torrent file.
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

	return tf, nil
}

// PieceHashes returns a slice of hex strings representing the SHA-1 hash of
// each piece in the torrent file.
func (tf *TorrentFile) PieceHashes() []string {
	hashes := []string{}

	for i := 0; i < len(tf.Info.Pieces); i += 20 {
		piece := []byte(tf.Info.Pieces[i : i+20])
		hashes = append(hashes, hex.EncodeToString(piece))
	}

	return hashes
}

// InfoHashHex returns the SHA-1 hash of the torrent info dictionary in hex format.
func (tf *TorrentFile) InfoHashHex() (string, error) {
	hash, err := tf.InfoHash()
	out := hex.EncodeToString([]byte(hash))
	return out, err
}

// InfoHashHex returns the SHA-1 hash of the torrent info dictionary in binary format.
func (tf *TorrentFile) InfoHash() (string, error) {
	h := sha1.New()
	err := bencode.Marshal(h, tf.Info)
	return string(h.Sum(nil)), err
}

func printInfoOutput(tf TorrentFile) {
	fmt.Printf("Tracker URL: %s\n", tf.Announce)
	fmt.Printf("Length: %d\n", tf.Info.Length)
	bc, _ := tf.InfoHashHex()
	fmt.Printf("Info Hash: %s\n", bc)
	fmt.Printf("Piece Length: %d\n", tf.Info.PieceLength)
	fmt.Println("Piece Hashes:")
	for _, hash := range tf.PieceHashes() {
		fmt.Println(hash)
	}
}
