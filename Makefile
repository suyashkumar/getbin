BINARY = getbin

.PHONY: build
build:
	dep ensure
	go build -o ${BINARY}

.PHONY: test
test:
	go test -v ./...

.PHONY: run
run:
	make build
	./${BINARY}

.PHONY: release
release:
	dep ensure
	GOOS=linux GOARCH=amd64 go build -o build/${BINARY}-linux-amd64 .;
	GOOS=darwin GOARCH=amd64 go build -o build/${BINARY}-darwin-amd64 .;
	GOOS=windows GOARCH=amd64 go build -o build/${BINARY}-windows-amd64.exe .;
	cd build; \
	tar -zcvf ${BINARY}-linux-amd64.tar.gz ${BINARY}-linux-amd64; \
	tar -zcvf ${BINARY}-darwin-amd64.tar.gz ${BINARY}-darwin-amd64; \
	zip -r ${BINARY}-windows-amd64.exe.zip ${BINARY}-windows-amd64.exe;
