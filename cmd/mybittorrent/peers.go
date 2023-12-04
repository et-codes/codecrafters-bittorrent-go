package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/jackpal/bencode-go"
)

const PeerID = "00112233445566778899" // Peer ID used for this client (20 bytes)

type PeerResponse struct {
	Complete    int    `bencode:"complete"`
	Incomplete  int    `bencode:"incomplete"`
	Interval    int    `bencode:"interval"`
	MinInterval int    `bencode:"min interval"`
	Peers       string `bencode:"peers"`
}

type HandshakeMessage struct {
	Protocol string // should be "BitTorrent protocol"
	Reserved []byte // should be {0, 0, 0, 0, 0, 0, 0, 0}
	InfoHash string // SHA-1 hash of torrent file info
	PeerID   string // ID of the peer
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

func (tf *TorrentFile) Handshake(peer string) (HandshakeMessage, error) {
	// Create the handshake message.
	message := newHandshakeMessage(tf.InfoHash)

	// Establish a connection with peer.
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		return HandshakeMessage{}, err
	}
	defer conn.Close()

	// Send the handshake.
	n, err := conn.Write(message)
	if err != nil {
		return HandshakeMessage{}, err
	}

	// Wait for the response.
	resp := make([]byte, n)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Read(resp)
	if err != nil {
		if err != io.EOF {
			return HandshakeMessage{}, err
		}
	}

	// Unmarshal the response into a HandshakeMessage.
	return parseHandshake(resp)
}

func parseHandshake(resp []byte) (HandshakeMessage, error) {
	result := HandshakeMessage{}
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

func (tf *TorrentFile) peerList() error {
	peers := []string{}
	pr, err := tf.discoverPeers()
	if err != nil {
		return err
	}

	for i := 0; i < len(pr.Peers); i += 6 {
		peer := pr.Peers[i : i+6]
		ip := peer[:4]
		portStr := []byte(peer[4:6])
		port := binary.BigEndian.Uint16(portStr)
		peerStr := fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], port)
		peers = append(peers, peerStr)
	}

	tf.Peers = peers

	return nil
}

func (tf *TorrentFile) discoverPeers() (PeerResponse, error) {
	peerResp := PeerResponse{}

	addr, err := peerRequestURL(tf.Announce, tf.InfoHash, tf.Info.Length)
	if err != nil {
		fmt.Println(err)
		return peerResp, err
	}

	res, err := http.Get(addr)
	if err != nil {
		fmt.Println(err)
		return peerResp, err
	}
	if res.StatusCode != http.StatusOK {
		fmt.Printf("Response code %d received.\n", res.StatusCode)
		return peerResp, err
	}

	err = bencode.Unmarshal(res.Body, &peerResp)
	if err != nil {
		fmt.Println(err)
		return peerResp, err

	}
	res.Body.Close()

	return peerResp, nil
}

func peerRequestURL(rawURL string, infoHash string, infoLength int) (string, error) {
	addr, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	values := addr.Query()
	values.Add("info_hash", infoHash)
	values.Add("peer_id", PeerID)
	values.Add("port", "6881")
	values.Add("uploaded", "0")
	values.Add("downloaded", "0")
	values.Add("left", fmt.Sprint(infoLength))
	values.Add("compact", "1")

	addr.RawQuery = values.Encode()

	return addr.String(), nil
}

func (tf *TorrentFile) PrintPeers() {
	for _, peer := range tf.Peers {
		fmt.Println(peer)
	}
}

func PrintHandshake(handshake HandshakeMessage) {
	fmt.Printf("Peer ID: %x\n", handshake.PeerID)
}
