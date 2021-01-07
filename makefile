SRC = $(shell find . -name 'docstring.go' -prune -o -name '*.go' -print)

build: lf lf.1

lf: $(SRC)
	CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-s -w" .

docstring.go: doc.go
	gen/docstring.sh

lf.1: docstring.go
	gen/man.sh

.PHONY: install clean
install: lf
	mv lf ~/.local/sbin/lf

clean:
	rm -f lf lf.1
