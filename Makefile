.PHONEY: build clean

build:  dist/deinde_darwin_amd64 \
	dist/deinde_linux_amd64 \
	dist/deinde_freebsd_amd64 \
	dist/deinde_openbsd_amd64

dist/deinde_darwin_amd64 dist/deinde_linux_amd64 deinde_freebsd_amd64 dist/deinde_openbsd_amd64: deinde.go
	go get github.com/mitchellh/gox
	gox -osarch="darwin/amd64 linux/amd64 freebsd/amd64 openbsd/amd64" -output="dist/deinde_{{.OS}}_{{.Arch}}"

clean:
	rm -rf dist
