.PHONY: test clean

# Jalankan semua unit test dengan verbose output
test:
	go test -v ./...

# Bersihkan file cache test build
clean:
	go clean -testcache
