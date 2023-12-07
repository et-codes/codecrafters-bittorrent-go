package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/jackpal/bencode-go"
)

type Client struct {
	Announce      string      // URL of the announce server
	Info          TorrentInfo // Torrent information
	InfoHash      string      // SHA-1 hash of the TorrentInfo data
	Peers         []string    // List of peer IP addresses
	ConnectedPeer int         // Index of the currently connected peer (-1 means none)
	PieceHashes   []string    // SHA-1 hashes of Pieces
}

type TorrentInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

// NewClient reads a torrent file and populates the a Client struct.
func NewClient(path string) (Client, error) {
	c := Client{
		ConnectedPeer: -1,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = decodeTorrentFile(string(data), &c)
	if err != nil {
		return c, err
	}

	infoHash, err := hashInfo(c.Info)
	if err != nil {
		return c, err
	}
	c.InfoHash = infoHash

	c.PieceHashes = hashPieces(c.Info.Pieces)

	err = c.peerList()
	if err != nil {
		return c, err
	}

	return c, nil
}

// Connect connects the client to the peer in Peers[peerIndex].
func (c *Client) Connect(peerIndex int) (net.Conn, error) {
	conn, err := net.Dial("tcp", c.Peers[peerIndex])
	if err != nil {
		return nil, err
	}
	c.ConnectedPeer = peerIndex
	return conn, nil
}

func (c *Client) Close(conn net.Conn) {
	conn.Close()
	c.ConnectedPeer = -1
}

func Handshake(conn io.ReadWriter, infoHash string) (Peer, error) {
	// Create the handshake message.
	message := newHandshakeMessage(infoHash)

	logger.Debug("Sending handshake...")
	n, err := conn.Write(message)
	if err != nil {
		return Peer{}, err
	}

	// Wait for the response.
	resp := make([]byte, n)
	_, err = conn.Read(resp)
	if err != nil {
		if err != io.EOF {
			return Peer{}, err
		}
	}

	peer, err := parseHandshake(resp)
	if err != nil {
		return peer, err
	}
	logger.Debug("Handshake returned from peer %x.\n", peer.PeerID)

	return peer, nil
}

func (c *Client) PrintInfo() {
	fmt.Printf("Tracker URL: %s\n", c.Announce)
	fmt.Printf("Length: %d\n", c.Info.Length)
	fmt.Printf("Info Hash: %s\n", infoHashHex(c.InfoHash))
	fmt.Printf("Piece Length: %d\n", c.Info.PieceLength)
	fmt.Println("Piece Hashes:")
	for _, hash := range c.PieceHashes {
		fmt.Println(hash)
	}
}

func PrintHandshake(handshake Peer) {
	fmt.Printf("Peer ID: %x\n", handshake.PeerID)
}

// hashPieces generates a slice of hex strings representing the SHA-1 hash of
// each piece in the torrent file.
func hashPieces(pieces string) []string {
	hashes := []string{}

	for i := 0; i < len(pieces); i += 20 {
		piece := []byte(pieces[i : i+20])
		hashes = append(hashes, hex.EncodeToString(piece))
	}

	return hashes
}

// infoHashHex returns the SHA-1 hash of the torrent info dictionary in hex format.
func infoHashHex(infoHash string) string {
	return hex.EncodeToString([]byte(infoHash))
}

// hashInfo calculates the SHA-1 hash of the torrent info dictionary in binary
// format and stores it in the InfoHash field of the Client struct.
func hashInfo(info TorrentInfo) (string, error) {
	h := sha1.New()
	err := bencode.Marshal(h, info)
	if err != nil {
		return "", err
	}
	infoHash := string(h.Sum(nil))
	return infoHash, nil
}

func newHandshakeMessage(infoHash string) []byte {
	protocolLength := byte(19)
	protocol := []byte("BitTorrent protocol")
	reserved := make([]byte, 8)

	message := append([]byte{protocolLength}, protocol...)
	message = append(message, reserved...)
	message = append(message, []byte(infoHash)...)
	message = append(message, []byte(PeerID)...)

	return message
}

func parseHandshake(resp []byte) (Peer, error) {
	result := Peer{}
	if len(resp) != 68 {
		return result, fmt.Errorf("expect response length 68, got %d", len(resp))
	}

	// Byte 0 should be 19, the length of the following protocol string
	result.Protocol = string(resp[1:20])  // 19 bytes
	result.Reserved = resp[20:28]         // 8 bytes
	result.InfoHash = string(resp[28:48]) // 20 bytes
	result.PeerID = string(resp[48:])     // 20 bytes

	return result, nil
}
