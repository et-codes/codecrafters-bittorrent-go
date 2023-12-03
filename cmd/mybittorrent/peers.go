package main

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jackpal/bencode-go"
)

type PeerResponse struct {
	Complete    int    `bencode:"complete"`
	Incomplete  int    `bencode:"incomplete"`
	Interval    int    `bencode:"interval"`
	MinInterval int    `bencode:"min interval"`
	Peers       string `bencode:"peers"`
}

type PeerList []string

func Peers(path string) error {
	tf, err := NewTorrentFile(path)
	if err != nil {
		return err
	}
	peers, err := tf.getPeerList()
	if err != nil {
		return err
	}
	printPeersOutput(peers)
	return nil
}

func (tf *TorrentFile) getPeerList() (PeerList, error) {
	peers := PeerList{}
	pr, err := tf.discoverPeers()
	if err != nil {
		return peers, err
	}

	for i := 0; i < len(pr.Peers); i += 6 {
		peer := pr.Peers[i : i+6]
		ip := peer[:4]
		portStr := []byte(peer[4:6])
		port := binary.BigEndian.Uint16(portStr)
		peerStr := fmt.Sprintf(
			"%d.%d.%d.%d:%d",
			ip[0],
			ip[1],
			ip[2],
			ip[3],
			port,
		)
		peers = append(peers, peerStr)
	}

	return peers, nil
}

func (tf *TorrentFile) discoverPeers() (PeerResponse, error) {
	peerResp := PeerResponse{}

	infoHash, err := tf.InfoHash()
	if err != nil {
		fmt.Println(err)
		return peerResp, err
	}

	addr, err := buildPeerRequestURL(tf.Announce, infoHash, tf.Info.Length)
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

func buildPeerRequestURL(rawURL string, infoHash string, infoLength int) (string, error) {
	addr, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	values := addr.Query()
	values.Add("info_hash", infoHash)
	values.Add("peer_id", "00112233445566778899")
	values.Add("port", "6881")
	values.Add("uploaded", "0")
	values.Add("downloaded", "0")
	values.Add("left", fmt.Sprint(infoLength))
	values.Add("compact", "1")

	addr.RawQuery = values.Encode()

	return addr.String(), nil
}

func printPeersOutput(peers PeerList) {
	for _, peer := range peers {
		fmt.Println(peer)
	}
}
