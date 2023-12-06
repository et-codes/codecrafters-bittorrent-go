package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/jackpal/bencode-go"
)

type TorrentInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

// PieceHashes generates a slice of hex strings representing the SHA-1 hash of
// each piece in the torrent file.
func (tf *Client) HashPieces() {
	hashes := []string{}

	for i := 0; i < len(tf.Info.Pieces); i += 20 {
		piece := []byte(tf.Info.Pieces[i : i+20])
		hashes = append(hashes, hex.EncodeToString(piece))
	}

	tf.PieceHashes = hashes
}

// InfoHashHex returns the SHA-1 hash of the torrent info dictionary in hex format.
func (tf *Client) InfoHashHex() string {
	out := hex.EncodeToString([]byte(tf.InfoHash))
	return out
}

// HashInfo calculates the SHA-1 hash of the torrent info dictionary in binary format.
func (tf *Client) HashInfo() error {
	h := sha1.New()
	err := bencode.Marshal(h, tf.Info)
	if err != nil {
		return err
	}
	tf.InfoHash = string(h.Sum(nil))
	return nil
}

func (tf *Client) PrintInfo() {
	fmt.Printf("Tracker URL: %s\n", tf.Announce)
	fmt.Printf("Length: %d\n", tf.Info.Length)
	fmt.Printf("Info Hash: %s\n", tf.InfoHashHex())
	fmt.Printf("Piece Length: %d\n", tf.Info.PieceLength)
	fmt.Println("Piece Hashes:")
	for _, hash := range tf.PieceHashes {
		fmt.Println(hash)
	}
}
