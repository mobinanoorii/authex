PACKAGES="./..."
# build paramters 
BUILD_FOLDER = dist
APP_VERSION = $(git describe --tags --always)

###############################################################################
###                           Basic Golang Commands                         ###
###############################################################################

all: install

install: go.sum
	go install main.go -o $(BUILD_FOLDER)/authex 

install-debug: go.sum
	go build -gcflags="all=-N -l" -o $(BUILD_FOLDER)/authex main.go

build: clean
	@echo build binary to $(BUILD_FOLDER)
	goreleaser build --single-target --config .github/.goreleaser.yaml --snapshot --rm-dist
	@echo done

clean:
	@echo clean build folder $(BUILD_FOLDER)
	rm -rf $(BUILD_FOLDER)
	@echo done

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify

test:
	@go test -mod=readonly $(PACKAGES) -cover -race

lint:
	@echo "--> Running linter"
	@golangci-lint run --config .github/.golangci.yaml
	@go mod verify

swagger-gen:
	@echo "installing deps"
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "generating swagger"
	swag init --dir web -g server.go
	@echo "dons	"


###############################################################################
###                                CI / CD                                  ###
###############################################################################

# TODO: running this with -race options causes problems in the cli tests
test-ci:
	go test -coverprofile=coverage.txt -covermode=atomic -mod=readonly $(PACKAGES)
