GO_TEST?=go test ./...
SWIFT_TEST?=cd analyzer-swift && swift test

.PHONY: test go-test swift-test fixturegen

test: go-test swift-test

go-test:
	$(GO_TEST)

swift-test:
	$(SWIFT_TEST)

fixturegen:
	go run ./cmd/fixturegen --out ./testdata/audio
