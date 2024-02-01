STOCK_BINARY=stockclient
TGTRETURN_BINARY=targetreturn
CURRETURNS_BINARY=currentreturn

all: build test

build:
	go build -o ./bin/${STOCK_BINARY} ./cmd/stockClient/stockClient.go
	go build -o ./bin/${TGTRETURN_BINARY} ./cmd/targetReturn/targetAnnualizedReturn.go
	go build -o ./bin/${CURRETURNS_BINARY} ./cmd/currentReturn/currentAnnualizedReturn.go

release:
	# Build Stock Client
	GOOS=windows GOARCH=amd64 go build -o ./bin/${STOCK_BINARY}-x64.exe ./cmd/stockClient/stockClient.go
	GOOS=windows GOARCH=386 go build -o ./bin/${STOCK_BINARY}.exe ./cmd/stockClient/stockClient.go
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${STOCK_BINARY}-mac-amd64 ./cmd/stockClient/stockClient.go
	GOOS=darwin GOARCH=arm64 go build -o ./bin/${STOCK_BINARY}-mac-arm64 ./cmd/stockClient/stockClient.go
	GOOS=linux GOARCH=amd64 go build -o ./bin/${STOCK_BINARY}-linux-amd64 ./cmd/stockClient/stockClient.go
	GOOS=linux GOARCH=arm64 go build -o ./bin/${STOCK_BINARY}-linux-arm64 ./cmd/stockClient/stockClient.go
	# Build Target Annualized Returns
	GOOS=windows GOARCH=amd64 go build -o ./bin/${TGTRETURN_BINARY}-x64.exe ./cmd/targetReturn/targetAnnualizedReturn.go
	GOOS=windows GOARCH=386 go build -o ./bin/${TGTRETURN_BINARY}.exe ./cmd/targetReturn/targetAnnualizedReturn.go
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${TGTRETURN_BINARY}-mac-amd64 ./cmd/targetReturn/targetAnnualizedReturn.go
	GOOS=darwin GOARCH=arm64 go build -o ./bin/${TGTRETURN_BINARY}-mac-arm64 ./cmd/targetReturn/targetAnnualizedReturn.go
	GOOS=linux GOARCH=amd64 go build -o ./bin/${TGTRETURN_BINARY}-linux-amd64 ./cmd/targetReturn/targetAnnualizedReturn.go
	GOOS=linux GOARCH=arm64 go build -o ./bin/${TGTRETURN_BINARY}-linux-arm64 ./cmd/targetReturn/targetAnnualizedReturn.go
	# Build Current Annualized Returns
	GOOS=windows GOARCH=amd64 go build -o ./bin/${CURRETURNS_BINARY}-x64.exe ./cmd/currentReturn/currentAnnualizedReturn.go
	GOOS=windows GOARCH=386 go build -o ./bin/${CURRETURNS_BINARY}.exe ./cmd/currentReturn/currentAnnualizedReturn.go
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${CURRETURNS_BINARY}-mac-amd64 ./cmd/currentReturn/currentAnnualizedReturn.go
	GOOS=darwin GOARCH=arm64 go build -o ./bin/${CURRETURNS_BINARY}-mac-arm64 ./cmd/currentReturn/currentAnnualizedReturn.go
	GOOS=linux GOARCH=amd64 go build -o ./bin/${CURRETURNS_BINARY}-linux-amd64 ./cmd/currentReturn/currentAnnualizedReturn.go
	GOOS=linux GOARCH=arm64 go build -o ./bin/${CURRETURNS_BINARY}-linux-arm64 ./cmd/currentReturn/currentAnnualizedReturn.go
	# Compress files
	zip bin/portfoliotools.windows-x64.zip bin/stockclient-x64.exe bin/currentreturn-x64.exe bin/targetreturn-x64.exe
	zip portfoliotools.windows.zip bin/stockclient.exe bin/currentreturn.exe bin/targetreturn.exe
	tar czvf bin/portfoliotools.mac-amd64.tar.gz bin/currentreturn-mac-amd64 bin/stockclient-mac-amd64 bin/targetreturn-mac-amd64
	tar czvf bin/portfoliotools.mac-arm64.tar.gz bin/currentreturn-mac-arm64 bin/stockclient-mac-arm64 bin/targetreturn-mac-arm64
	tar czvf bin/portfoliotools.linux-amd64.tar.gz bin/currentreturn-linux-amd64 bin/stockclient-linux-amd64 bin/targetreturn-linux-amd64
	tar czvf bin/portfoliotools.linux-arm64.tar.gz bin/currentreturn-linux-arm64 bin/stockclient-linux-arm64 bin/targetreturn-linux-arm64

test:
	go test -v ./pkg/

clean:
	go clean
	rm ./bin/*
