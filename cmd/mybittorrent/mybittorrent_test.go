package main

import (
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
	tf, err := NewTorrentFile("../../sample.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}

	tests := map[string]struct {
		got  interface{}
		want interface{}
	}{
		"tracker URL": {tf.Announce,
			"http://bittorrent-test-tracker.codecrafters.io/announce"},
		"length": {tf.Info.Length, 92063},
		"info hash": {hex.EncodeToString([]byte(tf.InfoHash)),
			"d69f91e6b2ae4c542468d1073a71d4ea13879a7f"},
		"piece hashes": {tf.PieceHashes, []string{
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
	tf, err := NewTorrentFile("../../sample.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}

	tests := map[string]struct {
		got  interface{}
		want interface{}
	}{
		"peer list": {tf.Peers, []string{
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
	tf, err := NewTorrentFile("../../sample.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}

	tests := map[string]struct {
		peer int
		want string
	}{
		"handshake peer 0": {0, "0000000000000000000000000000000000000000"},
		"handshake peer 1": {1, "2d524e302e302e302d5af5c2cf488815c4a2fa7f"},
		"handshake peer 2": {2, "2d524e302e302e302d0df778d11e39f854451c24"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tf.Handshake(tf.Peers[test.peer])
			if err != nil {
				t.Errorf(err.Error())
			}
			if hex.EncodeToString([]byte(got.PeerID)) != test.want {
				t.Errorf("got %q, wanted %q",
					hex.EncodeToString([]byte(got.PeerID)), test.want)
			}
		})
	}
}
