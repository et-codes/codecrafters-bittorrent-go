package main

type PeerMessage struct {
	Length  [4]byte // 4 bytes
	Type    byte    // 1 byte
	Payload []byte  // variable size
}

type RequestPayload struct {
	Index  int // piece index
	Begin  int // byte offset within the piece
	Length int // length of the block (16kb)
}

type PiecePayload struct {
	Index int    // piece index
	Begin int    // byte offset within the piece
	Block []byte // data for the piece
}

const blockLength = 16 * 1024 // 16kb

var messageTypes = map[string]byte{
	"choke":          byte(0), // no payload
	"unchoke":        byte(1), // no payload
	"interested":     byte(2), // no payload
	"not interested": byte(3), // no payload
	"have":           byte(4), // index just downloaded
	"bitfield":       byte(5), // indicates which pieces the peer has
	"request":        byte(6), // index, begin, and length
	"piece":          byte(7), // index, begin, and piece
	"cancel":         byte(8), // same payload as request message
}
