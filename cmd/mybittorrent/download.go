package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

func (tf *TorrentFile) DownloadPiece() error {
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
	resp := make([]byte, blockLength+13)
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
	_, err = conn.Read(resp)
	if err != nil {
		if err != io.EOF {
			return err
		}
	}

	length = binary.BigEndian.Uint32(resp[0:4])
	msgType = int(resp[4])
	msg = resp[5 : 5+length-1]
	fmt.Printf("UNCHOKE - Length: %d, Type: %d, Message: %b\n", length, msgType, msg)

	// Send 'request' message for each 16kb block, wait for corresponding 'piece' message
	payload := RequestPayload{
		Index:  0,
		Begin:  0,
		Length: blockLength,
	}
	req := requestPayloadToBytes(payload)
	_, err = conn.Write(req)
	if err != nil {
		return err
	}

	_, err = conn.Read(resp)
	if err != nil {
		if err != io.EOF {
			return err
		}
	}

	length = binary.BigEndian.Uint32(resp[0:4])
	msgType = int(resp[4])
	index := binary.BigEndian.Uint32(resp[5:9])
	offset := binary.BigEndian.Uint32(resp[9:13])
	msg = resp[13 : 13+length-9]
	fmt.Printf("PIECE - Length: %d, Type: %d, Index: %d, Offset: %d, Message: %s\n", length, msgType, index, offset, string(msg))

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
