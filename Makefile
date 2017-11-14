
GO_SOURCES=$(shell find ./ -type f -name '*.go')

.PHONY:
all: harbormaster

install: harbormaster
	go install .

harbormaster: $(GO_SOURCES)
	go build ./

.PHONY:
test:
	go test -v github.com/ilikeorangutans/harbormaster/...

