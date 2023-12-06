package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"time"
)

func (c *Client) DownloadPiece(conn io.ReadWriter, pieceIndex int, outputPath string) error {
	// Handshake and run preliminary protocol.
	err := initiateDownload(conn, pieceIndex, c.InfoHash)
	if err != nil {
		return err
	}

	// Calculate how many blocks are needed to fetch the entire piece.
	blocksRequired := int(math.Ceil(float64(c.Info.PieceLength) / float64(blockLength)))

	bytesReceived := 0
	pieceData := []byte{}

	// Send request message for each 16kb block, wait for corresponding piece message.
	for blockNum := 0; blockNum < blocksRequired; blockNum++ {
		blockSize := blockLength

		// Last block may be less than a full block length.
		if blockNum == blocksRequired-1 {
			blockSize = c.Info.PieceLength % blockLength
		}

		// Build request message.
		payload := requestPayloadToBytes(RequestPayload{
			Index:  uint32(pieceIndex),
			Offset: uint32(bytesReceived),
			Length: uint32(blockSize),
		})
		request := Message{
			Header:  MessageHeader{Type: msgRequest},
			Payload: payload,
		}

		// Send request message.
		log.Println("Sending request message...")
		err := sendMessage(conn, request)
		if err != nil {
			return err
		}

		// TODO figure out why we have to wait before reading message...
		time.Sleep(500 * time.Millisecond)

		// Get piece message.
		log.Println("Waiting for piece message...")
		piece, err := receiveMessage(conn, msgPiece)
		if err != nil {
			return err
		}
		index, offset, block := parsePiecePayload(piece)
		log.Printf("Piece message received, index %d, offset %d, block size %d.\n",
			index, offset, len(block))

		pieceData = append(pieceData, block...)
	}

	log.Println(string(pieceData))

	return nil
}

func initiateDownload(conn io.ReadWriter, pieceIndex int, infoHash string) error {
	// Execute handshake.
	_, err := Handshake(conn, infoHash)
	if err != nil {
		return err
	}

	// Get bitfield message.
	log.Println("Waiting for bitfield message...")
	bitfield, err := receiveMessage(conn, msgBitfield)
	if err != nil {
		return err
	}
	log.Printf("Bitfield message received: %+v\n", bitfield)

	// Make sure peer has the piece we're asking for.
	if !peerHasPiece(bitfield, pieceIndex) {
		return fmt.Errorf("peer does not have piece %d", pieceIndex)
	}

	log.Println("Sending interested message...")
	interested := Message{
		Header: MessageHeader{Type: msgInterested},
	}
	err = sendMessage(conn, interested)
	if err != nil {
		return err
	}

	// Get 'unchoke' message
	log.Println("Waiting for unchoke message...")
	unchoke, err := receiveMessage(conn, msgUnchoke)
	if err != nil {
		return err
	}
	log.Printf("Unchoke message received: %+v\n", unchoke)

	return nil
}

// peerHasPiece verifies whether the peer has the piece being requested.
func peerHasPiece(bitfield Message, pieceIndex int) bool {
	i := 0
	for _, bite := range bitfield.Payload {
		for j := 7; j >= 0; j-- {
			if bite>>j&1 == 1 && i == pieceIndex {
				return true
			}
			i++
		}
	}
	return false
}

// requestPayloadToBytes converts RequestPayload data into a byte slice to
// be added to the request message.
func requestPayloadToBytes(req RequestPayload) []byte {
	out := []byte{}

	// Piece index: 4 bytes
	pieceIndex := make([]byte, 4)
	binary.BigEndian.PutUint32(pieceIndex, req.Index)
	out = append(out, pieceIndex...)

	// Offset: 4 bytes
	offset := make([]byte, 4)
	binary.BigEndian.PutUint32(pieceIndex, req.Offset)
	out = append(out, offset...)

	// Length: 4 bytes (usually 16kb)
	blockLength := make([]byte, 4)
	binary.BigEndian.PutUint32(blockLength, req.Length)
	out = append(out, blockLength...)

	return out
}

// parsePiecePayload converts a piece message payload into index, offset,
// and block values.
func parsePiecePayload(piece Message) (index uint32, offset uint32, block []byte) {
	index = binary.BigEndian.Uint32(piece.Payload[0:4])
	offset = binary.BigEndian.Uint32(piece.Payload[4:8])
	if len(piece.Payload) > 8 {
		block = piece.Payload[8:]
	}
	return
}
