APP=cmd/sudoku-web
BIN?=bin
GOFLAGS?=
LDFLAGS?=

# runtime flags
SOLVER?=dlx           # dlx | backtrack
ADDR?=:8080
PERSIST?=./data

.PHONY: build run run-dlx run-backtrack test bench tidy cross cross-win cross-linux clean

build: | $(BIN)
	@echo "▶ building ($(SOLVER))..."
	@go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN)/sudoku-web ./$(APP)

run:
	@go run $(GOFLAGS) ./$(APP) -addr $(ADDR) -persist-path $(PERSIST) -solver $(SOLVER)

run-dlx:
	@$(MAKE) run SOLVER=dlx

run-backtrack:
	@$(MAKE) run SOLVER=backtrack

test:
	@go test ./...

bench:
	@go test -bench=. -benchmem ./...

tidy:
	@go mod tidy

cross: cross-win cross-linux

cross-win: | $(BIN)
	@echo "▶ cross-compiling Windows amd64"
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BIN)/sudoku-web-windows-amd64.exe ./$(APP)

cross-linux: | $(BIN)
	@echo "▶ cross-compiling Linux amd64"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BIN)/sudoku-web-linux-amd64 ./$(APP)

$(BIN):
	@mkdir -p $(BIN)

clean:
	@rm -rf $(BIN)