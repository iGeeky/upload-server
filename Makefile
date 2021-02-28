all: linux mac

linux: upload-server.linux
mac: ./dist/upload-server.mac

UTILS="github.com/iGeeky/open-account/pkg/baselib/utils"
GIT_FLAGS=-tags "jsoniter wechat_debug" -ldflags "-X $UTILS.GitCommit=`git rev-parse HEAD` -X $UTILS.GitBranch=`git rev-parse --abbrev-ref HEAD`" 

PROJ_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/)
BUILD=GOARCH=amd64 CGO_ENABLED=1 go build -o $@ ${GIT_FLAGS} $<
MAC_BUILD=GOOS=darwin $(BUILD)
LINUX_BUILD=GOOS=linux $(BUILD)


xfile.mac: pkg/xfile/c/xfile.c
	rm -f pkg/xfile/c/libxfile.Darwin.a
	cd pkg/xfile/c && make

xfile.linux: pkg/xfile/c/xfile.c
	rm -f pkg/xfile/c/libxfile.Linux.a
	$(DOCKER_CMD) /bin/bash -c "cd /app/pkg/xfile/c && make"

./dist/upload-server.mac: cmd/main.go
	$(MAC_BUILD)

./dist/upload-server.linux: cmd/main.go
	$(LINUX_BUILD)

./dist/%.mac: tools/%.go
	$(MAC_BUILD)

./dist/%.linux: tools/%.go
	$(LINUX_BUILD)

DOCKER_CMD=docker run --rm -ti --name=go-1.16-dev \
		-v ${GOPATH}:/go \
		-v ${PROJ_ROOT}:/app \
		golang:1.16
DOCKER_MAKE=rm -f dist/$@ && $(DOCKER_CMD) /bin/bash -c "cd /app && make ./dist/$@"

upload-server.linux:
	$(DOCKER_MAKE)

clean:
	rm -f ./dist/*
