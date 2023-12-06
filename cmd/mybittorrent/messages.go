package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

// MessageHeader represents the message length and type.
type MessageHeader struct {
	Length int
	Type   int
}

type Message struct {
	Header  MessageHeader // Header of the message
	Payload []byte        // Message payload
}

// RequestPayload is the payload for a Request message.
type RequestPayload struct {
	Index  int // piece index
	Offset int // byte offset within the piece
	Length int // length of the block (16kb)
}

// PiecePayload is the payload of a Piece message.
type PiecePayload struct {
	Index  [4]byte // piece index
	Offset [4]byte // byte offset within the piece
	Block  []byte  // data for the piece
}

// Length in bytes of the block size we are using in this app.
const blockLength = 16 * 1024 // 16kb

// Peer message types
const (
	msgChoke         = iota // 0 no payload
	msgUnchoke              // 1 no payload
	msgInterested           // 2 no payload
	msgNotInterested        // 3 no payload
	msgHave                 // 4 index just downloaded
	msgBitfield             // 5 indicates which pieces the peer has
	msgRequest              // 6 index, offest, and length
	msgPiece                // 7 index, offest, and piece index
	msgCancel               // 8 index, offest, and length
)

// receiveMessage reads a BitTorrent protocol response from the peer and
// returns its contents and an error.
func receiveMessage(conn io.ReadWriter) (Message, error) {
	// Get length header.
	resp := make([]byte, 4)
	_, err := conn.Read(resp)
	if err != nil {
		if err != io.EOF {
			return Message{}, err
		}
	}
	length := int(binary.BigEndian.Uint32(resp))

	// Get the type and payload.
	resp = make([]byte, length)
	n, err := conn.Read(resp)
	if err != nil {
		if err != io.EOF {
			return Message{}, err
		}
	}

	msgType := int(resp[0])
	payload := []byte{}
	if length > 1 {
		payload = resp[1:length]
	}

	// Make sure we received all of the message.
	if n < length {
		err = fmt.Errorf("only recieved %d bytes out of %d", n, length)
	} else {
		err = nil
	}

	return Message{
		Header: MessageHeader{
			Length: length,
			Type:   msgType,
		},
		Payload: payload,
	}, err
}

// sendMessage sends a message to the peer.
// func sendMessage(conn io.ReadWriter, msg Message) error {
// 	length := len(msg.Payload) + 1
// 	msgType := byte(msg.Header.Length)
// 	lengthPrefix := make([]byte, 4)
// 	binary.BigEndian.PutUint32(lengthPrefix, uint32(length))

// 	message := append(lengthPrefix, msgType)
// 	message = append(message, msg.Payload...)

// 	n, err := conn.Write(message)
// 	if err != nil {
// 		return err
// 	}
// 	if n != len(message) {
// 		return fmt.Errorf("expected to write %d bytes, only wrote %d",
// 			n, len(message))
// 	}

// 	return nil
// }
