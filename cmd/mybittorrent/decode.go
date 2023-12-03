package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackpal/bencode-go"
)

func Decode(bencodedValue string) error {
	decoded, err := decodeBencode(bencodedValue)
	if err != nil {
		return err
	}
	printDecodeOutput(decoded)
	return nil
}

func decodeBencode(bencodedString string) (interface{}, error) {
	reader := strings.NewReader(bencodedString)
	return bencode.Decode(reader)
}

func decodeTorrentFile(bencodedString string, out *TorrentFile) error {
	reader := strings.NewReader(bencodedString)
	return bencode.Unmarshal(reader, out)
}

func printDecodeOutput(decoded interface{}) {
	jsonOutput, _ := json.Marshal(decoded)
	fmt.Println(string(jsonOutput))
}
