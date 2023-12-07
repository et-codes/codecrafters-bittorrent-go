lint:
	@golangci-lint run

test: lint
	@go test -v ./...

test_all: test
	@codecrafters test

info:
	./your_bittorrent.sh info sample.torrent

peers:
	./your_bittorrent.sh peers sample.torrent

handshake:
	./your_bittorrent.sh handshake sample.torrent 178.62.85.20:51489

download_piece:
	./your_bittorrent.sh download_piece -o /tmp/test-piece-0 sample.torrent 0

download:
	./your_bittorrent.sh download -o /tmp/test.txt sample.torrent