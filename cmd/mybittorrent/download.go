package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
)

func (c *Client) DownloadPiece(conn io.ReadWriter, pieceIndex int, outputPath string) error {
	// Handshake and run preliminary protocol.
	err := initiateDownload(conn, pieceIndex, c.InfoHash)
	if err != nil {
		return err
	}

	// If last piece, calculate its size.
	var pieceLength int
	if pieceIndex == len(c.PieceHashes) {
		pieceLength = c.Info.Length % c.Info.PieceLength
		logger.Debug("Last piece length %d", pieceLength)
	} else {
		pieceLength = c.Info.PieceLength
	}

	// Calculate how many blocks are needed to fetch the entire piece.
	blocksRequired := int(math.Ceil(float64(pieceLength) / float64(blockLength)))

	pieceBytesReceived := 0
	piece := []byte{}

	logger.Debug("Requesting %d blocks to retreive piece of length %d...\n",
		blocksRequired, pieceLength)

	// Download each block.
	for blockNum := 1; blockNum <= blocksRequired; blockNum++ {
		blockBytesExpected := blockLength

		// Last block may be less than a full block length.
		if blockNum == blocksRequired {
			blockBytesExpected = pieceLength - pieceBytesReceived
		}

		block, err := downloadBlock(conn, pieceIndex, pieceBytesReceived, blockBytesExpected)
		if err != nil {
			return err
		}

		if block != nil {
			pieceBytesReceived += len(block)
			logger.Info("Block %d/%d received %d bytes.\n", blockNum, blocksRequired, len(block))
			piece = append(piece, block...)
		} else {
			logger.Debug("Received empty block before piece length, short piece.")
			break
		}
	}

	logger.Info("Piece download complete, downloaded %d/%d bytes.\n", pieceBytesReceived, c.Info.PieceLength)

	if !pieceIsValid(c.PieceHashes[pieceIndex], piece) {
		return fmt.Errorf("piece did not meet hash check")
	}
	logger.Info("Piece hash is valid.")

	err = savePiece(outputPath, piece)
	if err != nil {
		return err
	}

	return nil
}

// savePiece saves a piece to disk.
func savePiece(path string, piece []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := f.Write(piece)
	if err != nil {
		return err
	}

	if n != len(piece) {
		return fmt.Errorf("only wrote %d bytes, piece length %d", n, len(piece))
	}
	return nil
}

// pieceIsValid checks the hash of the piece received versus expected.
func pieceIsValid(pieceHash string, pieceData []byte) bool {
	h := sha1.New()

	_, err := h.Write(pieceData)
	if err != nil {
		logger.Error("Error hashing piece: %v", err)
		return false
	}

	hash := hex.EncodeToString(h.Sum(nil))

	return hash == pieceHash
}

func downloadBlock(conn io.ReadWriter, pieceIndex, offset, blockBytesExpected int) ([]byte, error) {
	// Build request message.
	payload := requestPayloadToBytes(RequestPayload{
		Index:  uint32(pieceIndex),
		Offset: uint32(offset),
		Length: uint32(blockBytesExpected),
	})
	request := Message{
		Header:  MessageHeader{Type: msgRequest},
		Payload: payload,
	}

	// Send request message.
	logger.Debug("Sending request message at offset %d...\n", offset)
	err := sendMessage(conn, request)
	if err != nil {
		return nil, err
	}

	// Get piece message.
	logger.Debug("Waiting for piece message...")
	piece, err := receiveMessage(conn, msgPiece)
	if err != nil {
		if piece.Header.Type == msgRejected {
			logger.Error("Request was rejected.")
		}
		return nil, err
	}
	i, o, block := parsePiecePayload(piece)
	logger.Debug("Piece message received: index %d, offset %d, block size %d.\n",
		i, o, len(block))

	return block, nil
}

func initiateDownload(conn io.ReadWriter, pieceIndex int, infoHash string) error {
	// Execute handshake.
	_, err := Handshake(conn, infoHash)
	if err != nil {
		return err
	}

	// Get bitfield message.
	logger.Debug("Waiting for bitfield message...")
	bitfield, err := receiveMessage(conn, msgBitfield)
	if err != nil {
		return err
	}
	logger.Debug("Bitfield message received: %+v\n", bitfield)

	// Make sure peer has the piece we're asking for.
	if !peerHasPiece(bitfield, pieceIndex) {
		return fmt.Errorf("peer does not have piece %d", pieceIndex)
	}

	logger.Debug("Sending interested message...")
	interested := Message{
		Header: MessageHeader{Type: msgInterested},
	}
	err = sendMessage(conn, interested)
	if err != nil {
		return err
	}

	// Get 'unchoke' message
	logger.Debug("Waiting for unchoke message...")
	unchoke, err := receiveMessage(conn, msgUnchoke)
	if err != nil {
		return err
	}
	logger.Debug("Unchoke message received: %+v\n", unchoke)

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
	binary.BigEndian.PutUint32(offset, req.Offset)
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
	if piece.Payload == nil {
		return 0, 0, nil
	}
	index = binary.BigEndian.Uint32(piece.Payload[0:4])
	offset = binary.BigEndian.Uint32(piece.Payload[4:8])
	if len(piece.Payload) > 8 {
		block = piece.Payload[8:]
	}
	return
}
