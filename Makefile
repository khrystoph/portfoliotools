STOCK_BINARY=stockclient
TGTRETURN_BINARY=targetreturn
CURRETURNS_BINARY=currentreturn

all: build test

build:
	go build -o ./bin/${STOCK_BINARY} ./cmd/stockClient/stockClient.go
	go build -o ./bin/${TGTRETURN_BINARY} ./cmd/targetReturn/targetAnnualizedReturn.go
	go build -o ./bin/${CURRETURNS_BINARY} ./cmd/currentReturn/currentAnnualizedReturn.go

test:
	go test -v ./pkg/

clean:
	go clean
	rm ./bin/${CURRETURNS_BINARY} ./bin/${TGTRETURN_BINARY} ./bin/${STOCK_BINARY}
