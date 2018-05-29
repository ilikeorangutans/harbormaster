
GO_SOURCES=$(shell find ./ -type f -name '*.go')
BINARY=harbormaster
MAC_OS_X_ZIP=harbormaster-macosx.zip

.PHONY:
all: $(BINARY)

.PHONY:
install: $(BINARY)
	go install .

$(BINARY): $(GO_SOURCES)
	go build ./

.PHONY:
test:
	go test -v github.com/ilikeorangutans/harbormaster/...

.PHONY:
clean:
	rm $(BINARY)

.PHONY:
dist: $(MAC_OS_X_ZIP)

$(MAC_OS_X_ZIP): $(BINARY)
	zip -9 harbormaster-macosx.zip harbormaster

