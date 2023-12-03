package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, error) {
	switch {
	case unicode.IsDigit(rune(bencodedString[0])):
		return decodeString(bencodedString)
	case bencodedString[0] == 'i': 
		return decodeInteger(bencodedString)
	default:
		return "", fmt.Errorf("only strings are supported at the moment")
	}
}

func decodeInteger(bencodedString string) (int, error) {
	bencodedString = strings.TrimPrefix(bencodedString, "i")
	bencodedString = strings.TrimSuffix(bencodedString, "e")

	integer, err := strconv.Atoi(bencodedString)
	if err != nil {
		return 0, err
	}

	return integer, nil
}

func decodeString(bencodedString string) (string, error) {
	var firstColonIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencodedString[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", err
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
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
