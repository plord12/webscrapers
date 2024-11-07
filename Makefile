include env
export

NAME=ha_ss
BINDIR=bin
SOURCES=$(wildcard *.go)
BINARIES=${BINDIR}/${NAME} ${BINDIR}/${NAME}-darwin-arm64 ${BINDIR}/${NAME}-linux-arm64

all: ${BINDIR} ${BINARIES}

${BINDIR}:
	mkdir -p ${BINDIR}

${BINDIR}/${NAME}: ${SOURCES}
	go build -o $@

${BINDIR}/${NAME}-darwin-arm64: ${SOURCES}
	GOARCH=arm64 GOOS=darwin go build -o $@

${BINDIR}/${NAME}-linux-arm64: ${SOURCES}
	GOARCH=arm64 GOOS=linux go build -o $@

test:
	${BINDIR}/${NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -url "$(TEST1_HA_URL)" -css "$(TEST1_HA_CSS)" -filename test1.png
	${BINDIR}/${NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -url "$(TEST2_HA_URL)" -css "$(TEST2_HA_CSS)" -filename test2.png
	${BINDIR}/${NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -url "$(TEST3_HA_URL)" -css "$(TEST3_HA_CSS)" -filename test3.png

testrest:
	${BINDIR}/${NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -restport 3500

clean:
	@go clean
	-@rm -rf ${BINDIR} test*.png 2>/dev/null || true

