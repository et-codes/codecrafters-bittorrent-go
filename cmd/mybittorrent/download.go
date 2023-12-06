package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"
)

func (c *Client) DownloadPiece(conn io.ReadWriter, pieceIndex int, outputPath string) error {
	// Handshake and run preliminary protocol.
	err := initiateDownload(conn, pieceIndex, c.InfoHash)
	if err != nil {
		return err
	}

	// Send 'request' message for each 16kb block, wait for corresponding 'piece' message
	blocksRequired := int(math.Ceil(float64(c.Info.PieceLength) / float64(blockLength)))
	bytesReceived := 0
	piece := []byte{}
	for b := 0; b < blocksRequired; b++ {
		blockSize := blockLength
		if b == blocksRequired-1 {
			blockSize = c.Info.PieceLength % blockLength
		}
		payload := RequestPayload{
			Index:  pieceIndex,
			Offset: bytesReceived,
			Length: blockSize,
		}

		// Send 'request' message
		req := requestPayloadToBytes(payload)
		_, err = conn.Write(req)
		if err != nil {
			return err
		}

		// Pause?
		time.Sleep(500 * time.Millisecond)

		// Read header
		header := make([]byte, 13)
		_, err = conn.Read(header)
		if err != nil {
			if err != io.EOF {
				return err
			}
		}

		length := binary.BigEndian.Uint32(header[:4])
		msgType := header[4]
		index := binary.BigEndian.Uint32(header[5:9])
		offset := binary.BigEndian.Uint32(header[9:13])

		// Expected block length is the length of the message minus the length
		// of the msgType, index, and offset (1 + 4 + 4)
		expectedBlockLength := int(length) - 9

		if msgType != byte(7) {
			fmt.Printf("wrong message id received, got %d\n", msgType)
			os.Exit(1)
		}
		fmt.Printf("PIECE - Length: %d, Type: %d, Index: %d, Offset: %d\n", length, msgType, index, offset)

		// Read block
		block := make([]byte, expectedBlockLength)
		for bytesReceived < expectedBlockLength {
			n, err := conn.Read(block[bytesReceived:])
			if err != nil {
				if err != io.EOF {
					return err
				}
			}
			bytesReceived += n
			fmt.Printf("===> Received %d bytes in block %d.\n", bytesReceived, b)
		}
		piece = append(piece, block...)
		// fmt.Println(string(block))
	}
	// fmt.Println("PIECE ==>", string(piece))
	fmt.Printf("Total bytes received: %d\n", len(piece))
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
	log.Printf("Bitfield received: %+v\n", bitfield)

	// Make sure peer has the piece we're asking for.
	if !peerHasPiece(bitfield, pieceIndex) {
		return fmt.Errorf("peer does not have piece %d", pieceIndex)
	}

	log.Println("Sending interested message...")
	interested := Message{
		Header: MessageHeader{
			Type: msgInterested,
		},
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
	log.Printf("Unchoke received: %+v\n", unchoke)

	return nil
}

func peerHasPiece(bitfield Message, piece int) bool {
	bits := ""
	for _, bit := range bitfield.Payload {
		if bit != 0 {
			bits += fmt.Sprintf("%b", bit)
		}
	}

	for i, b := range bits {
		if i == piece && b == '1' {
			return true
		}
	}
	return false
}

func requestPayloadToBytes(req RequestPayload) []byte {
	// Length prefix: 4 bytes
	out := []byte{}
	lengthPrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthPrefix, 13)
	out = append(out, lengthPrefix...)

	// Type: 1 byte
	out = append(out, msgRequest)

	// Piece index: 4 bytes
	pieceIndex := make([]byte, 4)
	binary.BigEndian.PutUint32(pieceIndex, uint32(req.Index))
	out = append(out, pieceIndex...)

	// Offset: 4 bytes
	offset := make([]byte, 4)
	binary.BigEndian.PutUint32(pieceIndex, uint32(req.Offset))
	out = append(out, offset...)

	// Length: 4 bytes (usually 16kb)
	blockLength := make([]byte, 4)
	binary.BigEndian.PutUint32(blockLength, uint32(req.Length))
	out = append(out, blockLength...)

	return out
}
