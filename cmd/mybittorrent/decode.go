package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackpal/bencode-go"
)

func Decode(bencodedValue string) (interface{}, error) {
	decoded, err := decodeBencode(bencodedValue)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func decodeBencode(bencodedString string) (interface{}, error) {
	reader := strings.NewReader(bencodedString)
	return bencode.Decode(reader)
}

func decodeTorrentFile(bencodedString string, out *Client) error {
	reader := strings.NewReader(bencodedString)
	return bencode.Unmarshal(reader, out)
}

func printDecodeOutput(decoded interface{}) {
	jsonOutput, _ := json.Marshal(decoded)
	fmt.Println(string(jsonOutput))
}
