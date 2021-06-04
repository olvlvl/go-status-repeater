SRC=./statusrepeater

lint:
	golangci-lint run

test:
	go test $(SRC)

test-coveralls:
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go test -v -covermode=count -coverprofile=coverage.out $(SRC)
