.PHONY: build run clean install

BINARY=lazymosh
DESTDIR=
PREFIX=/usr/local

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

install: build
	install -D -m755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)
