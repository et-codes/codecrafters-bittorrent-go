package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type Client struct {
	Announce    string      `bencode:"announce"`
	Info        TorrentInfo `bencode:"info"`
	InfoHash    string
	Peers       []string // List of peer IP addresses
	PieceHashes []string // SHA-1 hashes of Pieces
}

// NewClient reads a torrent file and populates the a Client struct.
func NewClient(path string) (Client, error) {
	c := Client{}

	data, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = decodeTorrentFile(string(data), &c)
	if err != nil {
		return c, err
	}

	err = c.HashInfo()
	if err != nil {
		return c, err
	}

	c.HashPieces()

	err = c.peerList()
	if err != nil {
		return c, err
	}

	return c, nil
}

func (c *Client) ConnectToPeer(index int) (net.Conn, error) {
	conn, err := net.Dial("tcp", c.Peers[index])
	if err != nil {
		return nil, err
	}
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	return conn, nil
}

func (c *Client) Handshake(conn io.ReadWriter) (Peer, error) {
	// Create the handshake message.
	message := newHandshakeMessage(c.InfoHash)

	log.Println("Sending handshake...")
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
	log.Printf("Handshake returned from peer %x.\n", peer.PeerID)

	return peer, nil
}

func PrintHandshake(handshake Peer) {
	fmt.Printf("Peer ID: %x\n", handshake.PeerID)
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
