package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func decodeBencode(bencodedString string) (interface{}, error) {
	result, _, err := doDecodeBencode(bencodedString)
	return result, err
}

func doDecodeBencode(bencodedString string) (interface{}, string, error) {
	var (
		result interface{}
		remainingString string
		err error
	)

	switch {
	case unicode.IsDigit(rune(bencodedString[0])):
		result, remainingString, err = decodeString(bencodedString)
	case bencodedString[0] == 'i': 
		result, remainingString, err = decodeInteger(bencodedString)
	case bencodedString[0] == 'l':
		result, remainingString, err = decodeList(bencodedString)
	default:
		return "", "", fmt.Errorf("unsupported type")
	}
	return result, remainingString, err
}

func decodeList(bencodedString string) ([]interface{}, string, error) {
	var remainingString string
	result := []interface{}{}

	bencodedString = strings.TrimPrefix(bencodedString, "l")
	
	// Find the position of the 'e' that ends the parent list
	var i int
	eCount := 1
	for i = 0; i < len(bencodedString); i++ {
		if bencodedString[i] == 'l' || bencodedString[i] == 'i' {
			eCount++ // embedded 'l' or 'i' means there's another 'e'
		} else if bencodedString[i] == 'e' {
			eCount--
		}
		if eCount == 0 {
			break
		}
	}

	// Build remaining string
	if i < len(bencodedString) - 1 {
		remainingString = bencodedString[i+1:]
	}
	bencodedString = bencodedString[:i]

	// Return if list is empty
	if bencodedString == "" {
		return result, "", nil
	}

	// Iterate through the elements inside the list
	for {
		value, remaining, err := doDecodeBencode(bencodedString)
		if err != nil {
			return nil, "", err
		}
		result = append(result, value)

		if remaining == "" {
			break
		}

		bencodedString = remaining
	}

	return result, remainingString, nil
}

func decodeInteger(bencodedString string) (int, string, error) {
	bencodedString = strings.TrimPrefix(bencodedString, "i")
	endIndex := strings.Index(bencodedString, "e")

	decodedInt, err := strconv.Atoi(bencodedString[:endIndex])
	if err != nil {
		return 0, "", err
	}

	remainingString := bencodedString[endIndex+1:]

	return decodedInt,	remainingString, nil
}

func decodeString(bencodedString string) (string, string, error) {
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
		return "", "", err
	}

	decodedString := bencodedString[firstColonIndex+1 : firstColonIndex+1+length]
	remainingString := bencodedString[firstColonIndex+1+length:]

	return decodedString, remainingString, nil
}
