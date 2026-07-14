APP     := Bauhaus
BUNDLE  := dist/$(APP).app
BIN     := bin/bauhaus
PKG     := ./cmd/bauhaus

.PHONY: all test build app install run clean fmt vet lint

all: test build

## test: unit tests with the race detector
test:
	go test -race ./...

## fmt/vet
fmt:
	gofmt -l -w .

vet:
	go vet ./...

lint: fmt vet test

## build: the plain binary
build:
	go build -o $(BIN) $(PKG)

## app: a real .app bundle (menu-bar app, LAN + Bonjour entitlements)
app: build icon
	rm -rf $(BUNDLE)
	mkdir -p $(BUNDLE)/Contents/MacOS $(BUNDLE)/Contents/Resources
	cp build/Info.plist $(BUNDLE)/Contents/Info.plist
	cp $(BIN)           $(BUNDLE)/Contents/MacOS/bauhaus
	cp build/AppIcon.icns $(BUNDLE)/Contents/Resources/AppIcon.icns
	# A stable signing identifier matters: Go's linker ad-hoc-signs every binary
	# with the identifier "a.out", so without this every Bauhaus build looks like
	# a different app to the firewall and to Local Network Privacy — which means a
	# fresh permission prompt on every rebuild.
	codesign --force --deep \
		--identifier dev.bauhaus.app \
		--sign - $(BUNDLE)
	@echo "built $(BUNDLE)"

## icon: generate AppIcon.icns from the source PNG
icon:
	@build/mkicon.sh

## install: put the app in /Applications and launch it
install: app
	rm -rf /Applications/$(APP).app
	cp -R $(BUNDLE) /Applications/
	open /Applications/$(APP).app
	@echo "Bauhaus is running in the menu bar."

## install-shared: let every macOS account on this Mac share one model cache.
##
## Without this, each user account keeps its own copy of every model — a 70B at
## 4-bit costs 40 GB twice. Bauhaus uses /Users/Shared/Bauhaus automatically once
## it exists and is writable.
##
## The setgid bit (the leading 2 in 2775) is the load-bearing part: it makes new
## files inherit the `staff` group, so a model downloaded by one account is
## writable by the next. /Users/Shared is world-writable and sticky by default,
## which is NOT enough on its own.
install-shared:
	sudo mkdir -p /Users/Shared/Bauhaus
	sudo chgrp -R staff /Users/Shared/Bauhaus
	sudo chmod -R 2775 /Users/Shared/Bauhaus
	@echo "Shared model cache ready at /Users/Shared/Bauhaus."
	@echo "Restart Bauhaus; every account on this Mac will now share one set of models."

## run: run headless in the foreground (for development)
run: build
	$(BIN) -headless

clean:
	rm -rf bin dist build/AppIcon.icns build/AppIcon.iconset
