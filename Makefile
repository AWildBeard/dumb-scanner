BUILD=go build
LDFLAGS=-X main.buildType=debug
DATE=$(shell date '+%s')
GOARCH=$(shell go env GOARCH)
OUT=$(shell basename $(PWD))

## Text coloring & styling
BOLD=\033[1m
UNDERLINE=\033[4m
HEADER=${BOLD}${UNDERLINE}

GREEN=\033[38;5;118m
RED=\033[38;5;196m
GREY=\033[38;5;250m

RESET=\033[m

.PHONY: release amd64

all: release build

l: lint
lint:
	@printf "${GREEN}${HEADER}Linting${RESET}\n"
	go vet ./...
amd64:
	$(eval GOARCH=amd64)
	@:
release:
	$(eval LDFLAGS=-w -s -X main.buildType=release)
	@:
b: build
build: clean
	$(eval LDFLAGS=${LDFLAGS} -X main.buildVersion=${DATE})
	@printf "${GREEN}${HEADER}Compiling for ${GOARCH}-${GOOS} to a file called '${OUT}'${RESET}\n"
	${BUILD} -p 1 -ldflags="${LDFLAGS}" -o ${OUT}
clean:
	@printf "${GREEN}${HEADER}Cleaning previous build${RESET}\n"
	rm -rf ${OUT}
