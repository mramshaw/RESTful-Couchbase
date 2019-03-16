GOPATH		:= /go
GOOS		:= linux
GOARCH		:= amd64

MAIN		:= restful_couchbase

.PHONY:		run, clean

all:		$(MAIN)
		@echo '$(MAIN)' has been compiled
		./$(MAIN)

$(MAIN):	build

# .go files are reformatted to conform to gofmt standards
fmt:
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) gofmt -d -e -s -w *.go
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) gofmt -d -e -s -w application/*.go
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) gofmt -d -e -s -w recipes/*.go
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) gofmt -d -e -s -w test/*.go

lint:		fmt
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) golint -set_exit_status *.go
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) golint -set_exit_status ./...

vet:		lint
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) go vet *.go
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) go vet application/*.go
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) go vet recipes/*.go
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) go vet test/*.go

test:		vet
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) go test -coverpkg .,application,recipes -coverprofile=coverage.txt -covermode=atomic -v . ./...
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) go tool cover -html=coverage.txt -o coverage.html

build:		test
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o $(MAIN) main.go

run:
		GOPATH=$(GOPATH) GOOS=$(GOOS) GOARCH=$(GOARCH) go run main.go

clean:
		rm ./$(MAIN)
