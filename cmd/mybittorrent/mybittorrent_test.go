package main

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := map[string]struct {
		encoded string
		want    interface{}
	}{
		"decode string":     {"5:hello", "hello"},
		"decode integer":    {"i52e", int64(52)},
		"decode list":       {"l5:helloi52ee", []interface{}{"hello", int64(52)}},
		"decode dictionary": {"d3:foo3:bar5:helloi52ee", map[string]interface{}{"foo": "bar", "hello": int64(52)}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := decodeBencode(test.encoded)
			if err != nil {
				t.Errorf("received error %v", err)
			}
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("got %v, wanted %v", got, test.want)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	c, err := NewClient("../../sample.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}

	tests := map[string]struct {
		got  interface{}
		want interface{}
	}{
		"tracker URL": {c.Announce,
			"http://bittorrent-test-tracker.codecrafters.io/announce"},
		"length": {c.Info.Length, 92063},
		"info hash": {hex.EncodeToString([]byte(c.InfoHash)),
			"d69f91e6b2ae4c542468d1073a71d4ea13879a7f"},
		"piece hashes": {c.PieceHashes, []string{
			"e876f67a2a8886e8f36b136726c30fa29703022d",
			"6e2275e604a0766656736e81ff10b55204ad8d35",
			"f00d937a0213df1982bc8d097227ad9e909acc17",
		}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Errorf("got %v, wanted %v", test.got, test.want)
			}
		})
	}
}

func TestPeers(t *testing.T) {
	c, err := NewClient("../../sample.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}

	tests := map[string]struct {
		got  interface{}
		want interface{}
	}{
		"peer list": {c.Peers, []string{
			"178.62.82.89:51470",
			"165.232.33.77:51467",
			"178.62.85.20:51489",
		}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Errorf("got %v, wanted %v", test.got, test.want)
			}
		})
	}
}

func TestHandshake(t *testing.T) {
	c, err := NewClient("../../sample.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}

	// Put a handshake response in the read buffer.
	buf := newHandshakeMessage(c.InfoHash)
	conn := bytes.NewBuffer(buf)

	tests := map[string]struct {
		want string
	}{
		"handshake has correct peer ID": {PeerID},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Handshake(conn, c.InfoHash)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.PeerID != test.want {
				t.Errorf("got %q, wanted %q",
					got.PeerID, test.want)
			}
		})
	}
}

func TestPeerHasPiece(t *testing.T) {
	tests := map[string]struct {
		bitfield []byte
		piece    int
		want     bool
	}{
		"detects bit 0":                {[]byte{byte(224)}, 0, true},
		"detects bit 1":                {[]byte{byte(224)}, 1, true},
		"detects bit 2":                {[]byte{byte(224)}, 2, true},
		"does not detect bit 3":        {[]byte{byte(224)}, 3, false},
		"works with more than one bit": {[]byte{byte(224), byte(1)}, 15, true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bitfield := Message{
				Payload: test.bitfield,
			}
			got := peerHasPiece(bitfield, test.piece)
			if got != test.want {
				t.Errorf("got %t, wanted %t",
					got, test.want)
			}
		})
	}
}

func TestDownloadPiece(t *testing.T) {}
