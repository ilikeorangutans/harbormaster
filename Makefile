
GO_SOURCES=$(shell find ./ -type f -name '*.go')
BINARY=harbormaster
MAC_OS_X_ZIP=harbormaster-macosx.zip

AZKABAN_VERSION=3.47.0
AZKABAN_ZIP=azkaban-$(AZKABAN_VERSION).tar.gz
AZKABAN_DOWNLOAD=https://github.com/azkaban/azkaban/archive/$(AZKABAN_VERSION).tar.gz
AZKABAN_DIR=azkaban-$(AZKABAN_VERSION)
AZKABAN_SOLO_SERVER=$(AZKABAN_DIR)/azkaban-solo-server/build/install/azkaban-solo-server

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
	-rm $(BINARY)
	-rm -rf dist

.PHONY:
dist-all: dist/harbormaster-darwin.zip dist/harbormaster-linux.zip dist/harbormaster-windows.zip

dist/harbormaster-%.zip: dist/%/harbormaster
	cd dist/$* && zip -9 ../harbormaster-$*.zip harbormaster

dist/%/harbormaster: $(GO_SOURCES)
	mkdir -vp dist/$*
	GOOS=$* GOARCH=amd64 go build -o dist/$*/harbormaster

.PHONY:
start-azkaban: $(AZKABAN_DIR)
	cd $(AZKABAN_DIR)/azkaban-solo-server/build/install/azkaban-solo-server && ./bin/start-solo.sh

.PHONY:
stop-azkaban: $(AZKABAN_DIR)
	cd $(AZKABAN_DIR)/azkaban-solo-server/build/install/azkaban-solo-server && ./bin/shutdown-solo.sh

$(AZKABAN_DIR): $(AZKABAN_ZIP)
	tar xzf $(AZKABAN_ZIP)
	cd $(AZKABAN_DIR) && ./gradlew distTar -x check


$(AZKABAN_ZIP):
	curl -L $(AZKABAN_DOWNLOAD) -o $(AZKABAN_ZIP)

