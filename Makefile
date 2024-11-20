# environment file containing test data
#
-include env
export

UTILS_NAME=utils
UTILS_SOURCE=${UTILS_NAME}/${UTILS_NAME}.go
HA_SS_NAME=ha_ss
HA_SS_SOURCE=${HA_SS_NAME}/${HA_SS_NAME}.go
AVIVA_NAME=aviva
AVIVA_SOURCE=${AVIVA_NAME}/${AVIVA_NAME}.go
AVIVAMYMONEY_NAME=avivamymoney
AVIVAMYMONEY_SOURCE=${AVIVAMYMONEY_NAME}/${AVIVAMYMONEY_NAME}.go
NUTMEG_NAME=nutmeg
NUTMEG_SOURCE=${NUTMEG_NAME}/${NUTMEG_NAME}.go
FUND_NAME=fund
FUND_SOURCE=${FUND_NAME}/${FUND_NAME}.go
MONEYFARM_NAME=moneyfarm
MONEYFARM_SOURCE=${MONEYFARM_NAME}/${MONEYFARM_NAME}.go
MONEYHUB_NAME=moneyhub
MONEYHUB_SOURCE=${MONEYHUB_NAME}/${MONEYHUB_NAME}.go
OCTOPUSWHEEL_NAME=octopuswheel
OCTOPUSWHEEL_SOURCE=${OCTOPUSWHEEL_NAME}/${OCTOPUSWHEEL_NAME}.go
MYCAUSEUK_NAME=mycauseuk
MYCAUSEUK_SOURCE=${MYCAUSEUK_NAME}/${MYCAUSEUK_NAME}.go

BINDIR=bin
BINARIES=${BINDIR}/${HA_SS_NAME} ${BINDIR}/${HA_SS_NAME}-linux-arm64
BINARIES+=${BINDIR}/${AVIVA_NAME} 
BINARIES+=${BINDIR}/${AVIVAMYMONEY_NAME} 
BINARIES+=${BINDIR}/${NUTMEG_NAME} 
BINARIES+=${BINDIR}/${FUND_NAME}
BINARIES+=${BINDIR}/${MONEYFARM_NAME} 
BINARIES+=${BINDIR}/${MONEYHUB_NAME} 
BINARIES+=${BINDIR}/${OCTOPUSWHEEL_NAME} 
BINARIES+=${BINDIR}/${MYCAUSEUK_NAME} 

all: ${BINDIR} ${BINARIES} otp

${BINDIR}:
	mkdir -p ${BINDIR}

otp:
	mkdir -p otp

# home assistant screenshots
#
${BINDIR}/${HA_SS_NAME}: ${HA_SS_SOURCE}
	go build -o $@ $<

${BINDIR}/${HA_SS_NAME}-linux-arm64: ${HA_SS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<

# aviva
#
${BINDIR}/${AVIVA_NAME}: ${AVIVA_SOURCE} ${UTILS_SOURCE}
	go build -o $@ $<

# aviva my money
#
${BINDIR}/${AVIVAMYMONEY_NAME}: ${AVIVAMYMONEY_SOURCE}
	go build -o $@ $<

# nutmeg
#
${BINDIR}/${NUTMEG_NAME}: ${NUTMEG_SOURCE}
	go build -o $@ $<

# fund
#
${BINDIR}/${FUND_NAME}: ${FUND_SOURCE}
	go build -o $@ $<

# moneyfarm
#
${BINDIR}/${MONEYFARM_NAME}: ${MONEYFARM_SOURCE}
	go build -o $@ $<

# moneyhub
#
${BINDIR}/${MONEYHUB_NAME}: ${MONEYHUB_SOURCE}
	go build -o $@ $<

# octopus wheel
#
${BINDIR}/${OCTOPUSWHEEL_NAME}: ${OCTOPUSWHEEL_SOURCE}
	go build -o $@ $<

# mycauseuk
#
${BINDIR}/${MYCAUSEUK_NAME}: ${MYCAUSEUK_SOURCE}
	go build -o $@ $<

test: testha testaviva testavivamymoney testnutmeg testfund testmoneyfarm testmoneyhub testoctopuswheel

testha: ${BINDIR}/${HA_SS_NAME}
	${BINDIR}/${HA_SS_NAME} -help
	${BINDIR}/${HA_SS_NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -url "$(TEST1_HA_URL)" -css "$(TEST1_HA_CSS)" -path test1.png
	${BINDIR}/${HA_SS_NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -url "$(TEST2_HA_URL)" -css "$(TEST2_HA_CSS)" -path test2.png
	${BINDIR}/${HA_SS_NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -url "$(TEST3_HA_URL)" -css "$(TEST3_HA_CSS)" -path test3.png

testrest:
	${BINDIR}/${HA_SS_NAME} -username "$(HA_USERNAME)" -password "$(HA_PASSWORD)" -restport 3500

testaviva: ${BINDIR}/${AVIVA_NAME}
	${BINDIR}/${AVIVA_NAME} -help
	${BINDIR}/${AVIVA_NAME} -username "$(AVIVA_USERNAME)" -password "$(AVIVA_PASSWORD)" -otpcleancommand "$(AVIVA_OTPCLEANCOMMAND)"  -otpcommand "$(AVIVA_OTPCOMMAND)"

testavivaselenium:
	OTP_CLEANCOMMAND="$(AVIVA_OTPCLEANCOMMAND)" OTP_COMMAND="$(AVIVA_OTPCOMMAND)" python3 ${AVIVA_NAME}/${AVIVA_NAME}.py

testavivamymoney: ${BINDIR}/${AVIVAMYMONEY_NAME}
	${BINDIR}/${AVIVAMYMONEY_NAME} -help
	${BINDIR}/${AVIVAMYMONEY_NAME} -username "$(AVIVAMYMONEY_USERNAME)" -password "$(AVIVAMYMONEY_PASSWORD)"  -word "$(AVIVAMYMONEY_WORD)"

testavivamymoneyselenium:
	python3 ${AVIVAMYMONEY_NAME}/${AVIVAMYMONEY_NAME}.py

testnutmeg: ${BINDIR}/${NUTMEG_NAME}
	${BINDIR}/${NUTMEG_NAME} -help
	${BINDIR}/${NUTMEG_NAME} -username "$(TEST1_NUTMEG_USERNAME)" -password "$(TEST1_NUTMEG_PASSWORD)" -otpcleancommand "$(NUTMEG_OTPCLEANCOMMAND)" -otpcommand "$(NUTMEG_OTPCOMMAND)"
	${BINDIR}/${NUTMEG_NAME} -username "$(TEST2_NUTMEG_USERNAME)" -password "$(TEST2_NUTMEG_PASSWORD)" -otpcleancommand "$(NUTMEG_OTPCLEANCOMMAND)" -otpcommand "$(NUTMEG_OTPCOMMAND)"

testfund: ${BINDIR}/${FUND_NAME}
	${BINDIR}/${FUND_NAME} -help
	${BINDIR}/${FUND_NAME} -fund "$(TEST1_FUND)"
	${BINDIR}/${FUND_NAME} -fund "$(TEST2_FUND)"
	${BINDIR}/${FUND_NAME} -fund "$(TEST3_FUND)"

testmoneyfarm: ${BINDIR}/${MONEYFARM_NAME}
	${BINDIR}/${MONEYFARM_NAME} -help
	${BINDIR}/${MONEYFARM_NAME} -username "$(TEST1_MONEYFARM_USERNAME)" -password "$(TEST1_MONEYFARM_PASSWORD)" -otpcleancommand "$(MONEYFARM_OTPCLEANCOMMAND)" -otpcommand "$(MONEYFARM_OTPCOMMAND)"
	${BINDIR}/${MONEYFARM_NAME} -username "$(TEST2_MONEYFARM_USERNAME)" -password "$(TEST2_MONEYFARM_PASSWORD)" -otpcleancommand "$(MONEYFARM_OTPCLEANCOMMAND)" -otpcommand "$(MONEYFARM_OTPCOMMAND)"

testmoneyhub: ${BINDIR}/${MONEYHUB_NAME}
	${BINDIR}/${MONEYHUB_NAME} -help
	MONEYHUB_BALANCE=$(shell ${BINDIR}/${NUTMEG_NAME} -username "$(TEST1_NUTMEG_USERNAME)" -password "$(TEST1_NUTMEG_PASSWORD)" -otpcleancommand "$(NUTMEG_OTPCLEANCOMMAND)" -otpcommand "$(NUTMEG_OTPCOMMAND)"   -headless=false); \
	${BINDIR}/${MONEYHUB_NAME} -username "$(TEST1_MONEYHUB_USERNAME)" -password "$(TEST1_MONEYHUB_PASSWORD)" -account "$(TEST1_MONEYHUB_ACCOUNT)" -balance $$MONEYHUB_BALANCE

testoctopuswheel: ${BINDIR}/${OCTOPUSWHEEL_NAME}
	${BINDIR}/${OCTOPUSWHEEL_NAME} -help
	${BINDIR}/${OCTOPUSWHEEL_NAME} -username "$(TEST1_OCTOPUS_USERNAME)" -password "$(TEST1_OCTOPUS_PASSWORD)"

testmycauseuk: ${BINDIR}/${MYCAUSEUK_NAME}
	${BINDIR}/${MYCAUSEUK_NAME} -help
	${BINDIR}/${MYCAUSEUK_NAME} -username "$(TEST1_MYCAUSEUK_USERNAME)" -password "$(TEST1_MYCAUSEUK_PASSWORD)"

clean:
	@go clean
	-@rm -rf ${BINDIR} test*.png 2>/dev/null || true

