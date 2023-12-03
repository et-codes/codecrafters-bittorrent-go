package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackpal/bencode-go"
)

func decodeBencode(bencodedString string) (interface{}, error) {
	reader := strings.NewReader(bencodedString)
	return bencode.Decode(reader)
}

func decodeTorrentFile(bencodedString string, out *TorrentFile) error {
	reader := strings.NewReader(bencodedString)
	return bencode.Unmarshal(reader, out)
}

func hashTorrentInfo(data TorrentInfo) (string, error) {
	h := sha1.New()
	err := bencode.Marshal(h, data)
	out := fmt.Sprintf("%x", h.Sum(nil)) // convert to hex string
	return out, err
}

func printDecodeOutput(decoded interface{}) {
	jsonOutput, _ := json.Marshal(decoded)
	fmt.Println(string(jsonOutput))
}
