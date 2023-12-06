package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"time"
)

func (tf *TorrentFile) DownloadPiece(outputPath string) error {
	// Establish a connection with peer
	conn, err := net.Dial("tcp", tf.Peers[1])
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	// Execute handshake
	_, err = tf.Handshake(conn, tf.Peers[1])
	if err != nil {
		return err
	}

	// Wait for 'bitfield' message
	resp := make([]byte, blockLength+9)
	_, err = conn.Read(resp)
	if err != nil {
		if err != io.EOF {
			return err
		}
	}

	length := binary.BigEndian.Uint32(resp[0:4])
	msgType := int(resp[4])
	msg := resp[5 : 5+length-1]
	fmt.Printf("BITFIELD - Length: %d, Type: %d, Message: %b\n", length, msgType, msg)

	// Send an 'interested' message
	_, err = conn.Write([]byte{0, 0, 0, 1, messageTypes["interested"]})
	if err != nil {
		return err
	}

	// Receive 'unchoke' message back
	resp = make([]byte, 5)
	_, err = conn.Read(resp)
	if err != nil {
		if err != io.EOF {
			return err
		}
	}

	length = binary.BigEndian.Uint32(resp[0:4])
	msgType = int(resp[4])
	fmt.Printf("UNCHOKE - Length: %d, Type: %d\n", length, msgType)

	// Send 'request' message for each 16kb block, wait for corresponding 'piece' message
	blocksRequired := int(math.Ceil(float64(tf.Info.PieceLength) / float64(blockLength)))
	bytesReceived := 0
	pieceIndex := 0
	piece := []byte{}
	for b := 0; b < blocksRequired; b++ {
		blockSize := blockLength
		if b == blocksRequired-1 {
			blockSize = tf.Info.PieceLength % blockLength
		}
		payload := RequestPayload{
			Index:  pieceIndex,
			Begin:  bytesReceived,
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
		fmt.Println(string(block))
	}
	// fmt.Println("PIECE ==>", string(piece))
	fmt.Printf("Total bytes received: %d\n", len(piece))
	return nil
}

func requestPayloadToBytes(req RequestPayload) []byte {
	// Length prefix: 4 bytes
	out := []byte{}
	lengthPrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthPrefix, 13)
	out = append(out, lengthPrefix...)

	// Type: 1 byte
	out = append(out, messageTypes["request"])

	// Piece index: 4 bytes
	pieceIndex := make([]byte, 4)
	binary.BigEndian.PutUint32(pieceIndex, uint32(req.Index))
	out = append(out, pieceIndex...)

	// Offset: 4 bytes
	offset := make([]byte, 4)
	binary.BigEndian.PutUint32(pieceIndex, uint32(req.Begin))
	out = append(out, offset...)

	// Length: 4 bytes (usually 16kb)
	blockLength := make([]byte, 4)
	binary.BigEndian.PutUint32(blockLength, uint32(req.Length))
	out = append(out, blockLength...)

	return out
}
