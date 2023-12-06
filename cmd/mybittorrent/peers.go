package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/jackpal/bencode-go"
)

const PeerID = "00112233445566778899" // Peer ID used for this client (20 bytes)

type GetPeersResponse struct {
	Complete    int    `bencode:"complete"`
	Incomplete  int    `bencode:"incomplete"`
	Interval    int    `bencode:"interval"`
	MinInterval int    `bencode:"min interval"`
	Peers       string `bencode:"peers"`
}

type Peer struct {
	Protocol string // should be "BitTorrent protocol"
	Reserved []byte // should be {0, 0, 0, 0, 0, 0, 0, 0}
	InfoHash string // SHA-1 hash of torrent file info
	PeerID   string // ID of the peer
}

// peerList discovers peers and adds their IP addresses to the Peers
// field of the Client struct.
func (c *Client) peerList() error {
	peers := []string{}
	pr, err := c.discoverPeers()
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

	c.Peers = peers

	return nil
}

// discoverPeers gets a list of peers from the announce URL.
func (c *Client) discoverPeers() (GetPeersResponse, error) {
	peerResp := GetPeersResponse{}

	addr, err := peerRequestURL(c.Announce, c.InfoHash, c.Info.Length)
	if err != nil {
		log.Println(err)
		return peerResp, err
	}

	res, err := http.Get(addr)
	if err != nil {
		log.Println(err)
		return peerResp, err
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("Response code %d received.\n", res.StatusCode)
		return peerResp, err
	}

	err = bencode.Unmarshal(res.Body, &peerResp)
	if err != nil {
		log.Println(err)
		return peerResp, err
	}
	res.Body.Close()

	return peerResp, nil
}

func peerRequestURL(rawURL string, infoHash string, infoLength int) (string, error) {
	addr, err := url.Parse(rawURL)
	if err != nil {
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

func PrintPeers(peers []string) {
	for _, peer := range peers {
		fmt.Println(peer)
	}
}
