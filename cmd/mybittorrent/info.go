package main

import (
	"fmt"
	"os"
)

type TorrentFile struct {
	Announce string `bencode:"announce"`
	Info     struct {
		Length      int    `bencode:"length"`
		Name        string `bencode:"name"`
		PieceLength int    `bencode:"piece length"`
		Pieces      string `bencode:"pieces"`
	}
}

func parseTorrentFile(path string) (TorrentFile, error) {
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

func printInfoOutput(tf TorrentFile) {
	fmt.Printf("Tracker URL: %s\n", tf.Announce)
	fmt.Printf("Length: %d\n", tf.Info.Length)
}
