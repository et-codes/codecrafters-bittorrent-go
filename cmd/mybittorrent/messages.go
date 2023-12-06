package main

// MessageHeader represents the message length and type.
type MessageHeader struct {
	Length [4]byte // 4 bytes
	Type   byte    // 1 byte
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
