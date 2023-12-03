package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

func decodeBencode(bencodedString string) (interface{}, error) {
	reader := strings.NewReader(bencodedString)
	result, err := bencode.Decode(reader)
	return result, err
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
