include env
export

HA_SS_NAME=ha_ss
HA_SS_SOURCE=ha_ss/ha_ss.go

AVIVA_NAME=aviva
AVIVA_SOURCE=aviva/aviva.go
AVIVAMYMONEY_NAME=avivamymoney
AVIVAMYMONEY_SOURCE=avivamymoney/avivamymoney.go

BINDIR=bin
BINARIES=${BINDIR}/${HA_SS_NAME} ${BINDIR}/${HA_SS_NAME}-darwin-arm64 ${BINDIR}/${HA_SS_NAME}-linux-arm64
BINARIES+=${BINDIR}/${AVIVA_NAME} ${BINDIR}/${AVIVA_NAME}-darwin-arm64 ${BINDIR}/${AVIVA_NAME}-linux-arm64
BINARIES+=${BINDIR}/${AVIVAMYMONEY_NAME} ${BINDIR}/${AVIVAMYMONEY_NAME}-darwin-arm64 ${BINDIR}/${AVIVAMYMONEY_NAME}-linux-arm64

all: ${BINDIR} ${BINARIES}

${BINDIR}:
	mkdir -p ${BINDIR}

# home assistant screenshots
#
${BINDIR}/${HA_SS_NAME}: ${HA_SS_SOURCE}
	go build -o $@ $<

${BINDIR}/${HA_SS_NAME}-darwin-arm64: ${HA_SS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<

${BINDIR}/${HA_SS_NAME}-linux-arm64: ${HA_SS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<

# aviva
#
${BINDIR}/${AVIVA_NAME}: ${AVIVA_SOURCE}
	go build -o $@ $<

${BINDIR}/${AVIVA_NAME}-darwin-arm64: ${AVIVA_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<

${BINDIR}/${AVIVA_NAME}-linux-arm64: ${AVIVA_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<

# aviva my money
#
${BINDIR}/${AVIVAMYMONEY_NAME}: ${AVIVAMYMONEY_SOURCE}
	go build -o $@ $<

${BINDIR}/${AVIVAMYMONEY_NAME}-darwin-arm64: ${AVIVAMYMONEY_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<

${BINDIR}/${AVIVAMYMONEY_NAME}-linux-arm64: ${AVIVAMYMONEY_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<

test: testha testaviva

testha: ${BINDIR}/${HA_SS_NAME}
	${BINDIR}/${HA_SS_NAME} -help
	${BINDIR}/${HA_SS_NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -url "$(TEST1_HA_URL)" -css "$(TEST1_HA_CSS)" -path test1.png
	${BINDIR}/${HA_SS_NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -url "$(TEST2_HA_URL)" -css "$(TEST2_HA_CSS)" -path test2.png
	${BINDIR}/${HA_SS_NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -url "$(TEST3_HA_URL)" -css "$(TEST3_HA_CSS)" -path test3.png

testrest:
	${BINDIR}/${HA_SS_NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -restport 3500

testaviva: ${BINDIR}/${AVIVA_NAME}
	rm -rf otp
	mkdir -p otp
	${BINDIR}/${AVIVA_NAME} -help
	${BINDIR}/${AVIVA_NAME} -username "$(AVIVA_USERNAME)" -password "$(AVIVA_PASSWORD)" -otpcommand "$(AVIVA_OTPCOMMAND)"

testavivamymoney: ${BINDIR}/${AVIVAMYMONEY_NAME}
	rm -rf otp
	mkdir -p otp
	${BINDIR}/${AVIVAMYMONEY_NAME} -help
	${BINDIR}/${AVIVAMYMONEY_NAME} -username "$(AVIVAMYMONEY_USERNAME)" -password "$(AVIVAMYMONEY_PASSWORD)"  -word "$(AVIVAMYMONEY_WORD)"



clean:
	@go clean
	-@rm -rf ${BINDIR} test*.png 2>/dev/null || true

