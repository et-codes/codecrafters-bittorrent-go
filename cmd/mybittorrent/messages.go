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
	Index  uint32 // piece index
	Offset uint32 // byte offset within the piece
	Length uint32 // length of the block (16kb)
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
	msgRejected      = 16   // 16 request rejected by peer
)

// receiveMessage reads a BitTorrent protocol response from the peer and
// returns its contents and an error.
func receiveMessage(conn io.ReadWriter, expectedType int) (Message, error) {
	message := Message{}

	// Get message length.
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return message, err
	}

	length := int(binary.BigEndian.Uint32(header))
	message.Header.Length = length
	if length == 0 {
		return message, fmt.Errorf("message received has 0 length")
	}

	// Get message type.
	mt := make([]byte, 1)
	if _, err := io.ReadFull(conn, mt); err != nil {
		return message, err
	}

	msgType := int(mt[0])
	message.Header.Type = msgType

	if msgType != expectedType {
		return message,
			fmt.Errorf("expected message type %d, received %d", expectedType, msgType)
	}

	// Return now if there is no payload.
	if length == 1 {
		return message, nil
	}

	// Get the payload.
	payload := make([]byte, length-1)
	n, err := io.ReadFull(conn, payload)
	if err != nil {
		return message, err
	}
	message.Payload = payload

	logger.Debug("Received %d bytes.", n)

	return message, nil
}

// sendMessage sends a message to the peer.
func sendMessage(conn io.ReadWriter, msg Message) error {
	length := len(msg.Payload) + 1
	msgType := byte(msg.Header.Type)
	lengthPrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthPrefix, uint32(length))

	message := append(lengthPrefix, msgType)
	message = append(message, msg.Payload...)

	n, err := conn.Write(message)
	if err != nil {
		return err
	}
	if n != len(message) {
		return fmt.Errorf("expected to write %d bytes, only wrote %d",
			n, len(message))
	}

	return nil
}
