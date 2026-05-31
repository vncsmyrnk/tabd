SRCDIR = .
OUTPUT = $(SRCDIR)/build

GO_SRC = $(shell find . -type f -name '*.go')

TARGET = $(OUTPUT)/tabd

PREFIX ?= /usr
DESTDIR ?=

DATAROOTDIR = $(PREFIX)/share
DATADIR = $(DATAROOTDIR)
BINDIR = $(PREFIX)/bin
MANDIR = $(DATAROOTDIR)/man
MANDIR1 = $(MANDIR)/man1
ZSHDIR = $(DATAROOTDIR)/zsh

INSTALL ?= install
INSTALL_PROGRAM = $(INSTALL)
INSTALL_DATA = $(INSTALL) -m 644

GO ?= go
GO_FLAGS = -trimpath
GO_ENV ?= CGO_ENABLED=0
GO_LDFLAGS ?= -s -w

all: $(TARGET)

$(TARGET): $(GO_SRC)
	$(GO_ENV) $(GO) build \
		$(GO_FLAGS) \
		-ldflags="$(GO_LDFLAGS)" \
		-o $@ ./cmd/tabd/main.go

.PHONY: clean
clean:
	@rm -rf $(OUTPUT)

.PHONY: install
install: all
	$(INSTALL) -d $(DESTDIR)$(BINDIR)
	$(INSTALL_PROGRAM) $(TARGET) $(DESTDIR)$(BINDIR)

.PHONY: uninstall
uninstall:
	rm -rf $(DESTDIR)$(BINDIR)/tabd

.PHONY: check
check: $(GO_SRC)
	go test ./...

.PHONY: lint
lint: $(GO_SRC)
	golangci-lint run $(SRCDIR)/...
