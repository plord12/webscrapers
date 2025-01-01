# environment file containing test data
#
-include env
export

VERSION=v0.0.1-alpha

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
LAUNCH_NAME=launch
LAUNCH_SOURCE=${LAUNCH_NAME}/constants.go ${LAUNCH_NAME}/exec.go ${LAUNCH_NAME}/load-addons.go ${LAUNCH_NAME}/main.go ${LAUNCH_NAME}/procgroup-unix.go ${LAUNCH_NAME}/procgroup-win.go ${LAUNCH_NAME}/validate.go ${LAUNCH_NAME}/xpi-dl.go

# local local builds/tests
#
BINDIR=bin
BINARIES=${BINDIR}/${HA_SS_NAME}
BINARIES+=${BINDIR}/${LAUNCH_NAME} 
BINARIES+=${BINDIR}/${AVIVA_NAME} 
BINARIES+=${BINDIR}/${AVIVAMYMONEY_NAME} 
BINARIES+=${BINDIR}/${NUTMEG_NAME} 
BINARIES+=${BINDIR}/${FUND_NAME}
BINARIES+=${BINDIR}/${MONEYFARM_NAME} 
BINARIES+=${BINDIR}/${MONEYHUB_NAME} 
BINARIES+=${BINDIR}/${OCTOPUSWHEEL_NAME} 
BINARIES+=${BINDIR}/${MYCAUSEUK_NAME} 

LINUXARM64BINDIR=bin-linux-arm64
LINUXAMD64BINDIR=bin-linux-amd64
DARWINARM64BINDIR=bin-darwin-arm64
WINDOWSAMD64BINDIR=bin-windows-amd64

all: ${BINDIR} ${BINARIES} otp

${BINDIR}:
	mkdir -p ${BINDIR}

${LINUXARM64BINDIR}:
	mkdir -p ${LINUXARM64BINDIR}

otp:
	mkdir -p otp

release: release-ha_ss_addon release-launch-linux-arm64 release-launch-linux-amd64 release-launch-darwin-arm64 release-launch-windows-amd64 release-webscrapers-linux-arm64 release-webscrapers-linux-amd64 release-webscrapers-darwin-arm64 release-webscrapers-windows-amd64

release-ha_ss_addon: ${LINUXARM64BINDIR}/${HA_SS_NAME}  ${LINUXAMD64BINDIR}/${HA_SS_NAME}
	cp ${LINUXARM64BINDIR}/${HA_SS_NAME} ha_ss_addon/${HA_SS_NAME}-linux-arm64
	cp ${LINUXAMD64BINDIR}/${HA_SS_NAME} ha_ss_addon/${HA_SS_NAME}-linux-amd64
	zip ha_ss_addon-${VERSION}.zip ha_ss_addon/Dockerfile* ha_ss_addon/*yaml ha_ss_addon/run.sh ha_ss_addon/${HA_SS_NAME}-linux-arm64  ha_ss_addon/${HA_SS_NAME}-linux-amd64
	rm ha_ss_addon/${HA_SS_NAME}-linux-arm64  ha_ss_addon/${HA_SS_NAME}-linux-amd64

release-launch-linux-arm64: ${LINUXARM64BINDIR}/${LAUNCH_NAME}
	cd ${LINUXARM64BINDIR} && zip ../launch-linux-arm64-${VERSION}.zip $(notdir $^)

release-launch-linux-amd64: ${LINUXAMD64BINDIR}/${LAUNCH_NAME}
	cd ${LINUXAMD64BINDIR} && zip ../launch-linux-amd64-${VERSION}.zip $(notdir $^)

release-launch-darwin-arm64: ${DARWINARM64BINDIR}/${LAUNCH_NAME}
	cd ${DARWINARM64BINDIR} && zip ../launch-darwin-arm64-${VERSION}.zip $(notdir $^)

release-launch-windows-amd64: ${WINDOWSAMD64BINDIR}/${LAUNCH_NAME}.exe
	cd ${WINDOWSAMD64BINDIR} && zip ../launch-linux-windows-${VERSION}.zip $(notdir $^)

release-webscrapers-linux-arm64: ${LINUXARM64BINDIR}/${AVIVA_NAME} ${LINUXARM64BINDIR}/${AVIVAMYMONEY_NAME} ${LINUXARM64BINDIR}/${NUTMEG_NAME} ${LINUXARM64BINDIR}/${FUND_NAME} ${LINUXARM64BINDIR}/${MONEYFARM_NAME} ${LINUXARM64BINDIR}/${OCTOPUSWHEEL_NAME} ${LINUXARM64BINDIR}/${MYCAUSEUK_NAME}
	cd ${LINUXARM64BINDIR} && zip ../webscrapers-linux-arm64-${VERSION}.zip $(notdir $^)

release-webscrapers-linux-amd64: ${LINUXAMD64BINDIR}/${AVIVA_NAME} ${LINUXAMD64BINDIR}/${AVIVAMYMONEY_NAME} ${LINUXAMD64BINDIR}/${NUTMEG_NAME} ${LINUXAMD64BINDIR}/${FUND_NAME} ${LINUXAMD64BINDIR}/${MONEYFARM_NAME} ${LINUXAMD64BINDIR}/${OCTOPUSWHEEL_NAME} ${LINUXAMD64BINDIR}/${MYCAUSEUK_NAME}
	cd ${LINUXAMD64BINDIR} && zip ../webscrapers-linux-amd64-${VERSION}.zip $(notdir $^)

release-webscrapers-darwin-arm64: ${DARWINARM64BINDIR}/${AVIVA_NAME} ${DARWINARM64BINDIR}/${AVIVAMYMONEY_NAME} ${DARWINARM64BINDIR}/${NUTMEG_NAME} ${DARWINARM64BINDIR}/${FUND_NAME} ${DARWINARM64BINDIR}/${MONEYFARM_NAME} ${DARWINARM64BINDIR}/${OCTOPUSWHEEL_NAME} ${DARWINARM64BINDIR}/${MYCAUSEUK_NAME}
	cd ${DARWINARM64BINDIR} && zip ../webscrapers-darwin-arm64-${VERSION}.zip $(notdir $^)

release-webscrapers-windows-amd64: ${WINDOWSAMD64BINDIR}/${AVIVA_NAME}.exe ${WINDOWSAMD64BINDIR}/${AVIVAMYMONEY_NAME}.exe ${WINDOWSAMD64BINDIR}/${NUTMEG_NAME}.exe ${WINDOWSAMD64BINDIR}/${FUND_NAME}.exe ${WINDOWSAMD64BINDIR}/${MONEYFARM_NAME}.exe ${WINDOWSAMD64BINDIR}/${OCTOPUSWHEEL_NAME}.exe ${WINDOWSAMD64BINDIR}/${MYCAUSEUK_NAME}.exe
	cd ${WINDOWSAMD64BINDIR} && zip ../webscrapers-windows-amd64-${VERSION}.zip $(notdir $^)


# launcher
#
${BINDIR}/${LAUNCH_NAME}: ${LAUNCH_SOURCE}
	cd ${LAUNCH_NAME} && go build -o ../$@
${LINUXARM64BINDIR}/${LAUNCH_NAME}: ${LAUNCH_SOURCE}
	cd ${LAUNCH_NAME} && GOARCH=arm64 GOOS=linux go build -o ../$@
${LINUXAMD64BINDIR}/${LAUNCH_NAME}: ${LAUNCH_SOURCE}
	cd ${LAUNCH_NAME} && GOARCH=amd64 GOOS=linux go build -o ../$@
${DARWINARM64BINDIR}/${LAUNCH_NAME}: ${LAUNCH_SOURCE}
	cd ${LAUNCH_NAME} && GOARCH=arm64 GOOS=darwin go build -o ../$@
${WINDOWSAMD64BINDIR}/${LAUNCH_NAME}.exe: ${LAUNCH_SOURCE}
	cd ${LAUNCH_NAME} && GOARCH=amd64 GOOS=windows go build -o ../$@

# home assistant screenshots
#
${BINDIR}/${HA_SS_NAME}: ${HA_SS_SOURCE}
	go build -o $@ $<
${LINUXARM64BINDIR}/${HA_SS_NAME}: ${HA_SS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<
${LINUXAMD64BINDIR}/${HA_SS_NAME}: ${HA_SS_SOURCE}
	GOARCH=amd64 GOOS=linux go build -o $@ $<
${DARWINARM64BINDIR}/${HA_SS_NAME}: ${HA_SS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<

# aviva
#
${BINDIR}/${AVIVA_NAME}: ${AVIVA_SOURCE} ${UTILS_SOURCE}
	go build -o $@ $<
${LINUXARM64BINDIR}/${AVIVA_NAME}: ${AVIVA_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<
${LINUXAMD64BINDIR}/${AVIVA_NAME}: ${AVIVA_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=linux go build -o $@ $<
${DARWINARM64BINDIR}/${AVIVA_NAME}: ${AVIVA_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<
${WINDOWSAMD64BINDIR}/${AVIVA_NAME}.exe: ${AVIVA_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=windows go build -o $@ $<

# aviva my money
#
${BINDIR}/${AVIVAMYMONEY_NAME}: ${AVIVAMYMONEY_SOURCE} ${UTILS_SOURCE}
	go build -o $@ $<
${LINUXARM64BINDIR}/${AVIVAMYMONEY_NAME}: ${AVIVAMYMONEY_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<
${LINUXAMD64BINDIR}/${AVIVAMYMONEY_NAME}: ${AVIVAMYMONEY_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=linux go build -o $@ $<
${DARWINARM64BINDIR}/${AVIVAMYMONEY_NAME}: ${AVIVAMYMONEY_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<
${WINDOWSAMD64BINDIR}/${AVIVAMYMONEY_NAME}.exe: ${AVIVAMYMONEY_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=windows go build -o $@ $<

# nutmeg
#
${BINDIR}/${NUTMEG_NAME}: ${NUTMEG_SOURCE} ${UTILS_SOURCE}
	go build -o $@ $<
${LINUXARM64BINDIR}/${NUTMEG_NAME}: ${NUTMEG_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<
${LINUXAMD64BINDIR}/${NUTMEG_NAME}: ${NUTMEG_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=linux go build -o $@ $<
${DARWINARM64BINDIR}/${NUTMEG_NAME}: ${NUTMEG_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<
${WINDOWSAMD64BINDIR}/${NUTMEG_NAME}.exe: ${NUTMEG_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=windows go build -o $@ $<

# fund
#
${BINDIR}/${FUND_NAME}: ${FUND_SOURCE} ${UTILS_SOURCE}
	go build -o $@ $<
${LINUXARM64BINDIR}/${FUND_NAME}: ${FUND_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<
${LINUXAMD64BINDIR}/${FUND_NAME}: ${FUND_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=linux go build -o $@ $<
${DARWINARM64BINDIR}/${FUND_NAME}: ${FUND_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<
${WINDOWSAMD64BINDIR}/${FUND_NAME}.exe: ${FUND_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=windows go build -o $@ $<

# moneyfarm
#
${BINDIR}/${MONEYFARM_NAME}: ${MONEYFARM_SOURCE} ${UTILS_SOURCE}
	go build -o $@ $<
${LINUXARM64BINDIR}/${MONEYFARM_NAME}: ${MONEYFARM_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<
${LINUXAMD64BINDIR}/${MONEYFARM_NAME}: ${MONEYFARM_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=linux go build -o $@ $<
${DARWINARM64BINDIR}/${MONEYFARM_NAME}: ${MONEYFARM_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<
${WINDOWSAMD64BINDIR}/${MONEYFARM_NAME}.exe: ${MONEYFARM_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=windows go build -o $@ $<

# moneyhub
#
${BINDIR}/${MONEYHUB_NAME}: ${MONEYHUB_SOURCE} ${UTILS_SOURCE}
	go build -o $@ $<
${LINUXARM64BINDIR}/${MONEYHUB_NAME}: ${MONEYHUB_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<
${LINUXAMD64BINDIR}/${MONEYHUB_NAME}: ${MONEYHUB_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=linux go build -o $@ $<
${DARWINARM64BINDIR}/${MONEYHUB_NAME}: ${MONEYHUB_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<
${WINDOWSAMD64BINDIR}/${MONEYHUB_NAME}.exe: ${MONEYHUB_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=windows go build -o $@ $<

# octopus wheel
#
${BINDIR}/${OCTOPUSWHEEL_NAME}: ${OCTOPUSWHEEL_SOURCE} ${UTILS_SOURCE}
	go build -o $@ $<
${LINUXARM64BINDIR}/${OCTOPUSWHEEL_NAME}: ${OCTOPUSWHEEL_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<
${LINUXAMD64BINDIR}/${OCTOPUSWHEEL_NAME}: ${OCTOPUSWHEEL_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=linux go build -o $@ $<
${DARWINARM64BINDIR}/${OCTOPUSWHEEL_NAME}: ${OCTOPUSWHEEL_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<
${WINDOWSAMD64BINDIR}/${OCTOPUSWHEEL_NAME}.exe: ${OCTOPUSWHEEL_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=windows go build -o $@ $<

# mycauseuk
#
${BINDIR}/${MYCAUSEUK_NAME}: ${MYCAUSEUK_SOURCE} ${UTILS_SOURCE}
	go build -o $@ $<
${LINUXARM64BINDIR}/${MYCAUSEUK_NAME}: ${MYCAUSEUK_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=linux go build -o $@ $<
${LINUXAMD64BINDIR}/${MYCAUSEUK_NAME}: ${MYCAUSEUK_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=linux go build -o $@ $<
${DARWINARM64BINDIR}/${MYCAUSEUK_NAME}: ${MYCAUSEUK_SOURCE} ${UTILS_SOURCE}
	GOARCH=arm64 GOOS=darwin go build -o $@ $<
${WINDOWSAMD64BINDIR}/${MYCAUSEUK_NAME}.exe: ${MYCAUSEUK_SOURCE} ${UTILS_SOURCE}
	GOARCH=amd64 GOOS=windows go build -o $@ $<

test: testha testaviva testavivamymoney testnutmeg testfund testmoneyfarm testmoneyhub testoctopuswheel

testha: ${BINDIR}/${HA_SS_NAME}
	${BINDIR}/${HA_SS_NAME} --help
	${BINDIR}/${HA_SS_NAME} --username "$(HA_USERNAME)" --password "$(HA_PASSWORD)" --url "$(TEST1_HA_URL)" --css "$(TEST1_HA_CSS)" --path test1.png
	${BINDIR}/${HA_SS_NAME} --username "$(HA_USERNAME)" --password "$(HA_PASSWORD)" --url "$(TEST2_HA_URL)" --css "$(TEST2_HA_CSS)" --path test2.png
	${BINDIR}/${HA_SS_NAME} --username "$(HA_USERNAME)" --password "$(HA_PASSWORD)" --url "$(TEST3_HA_URL)" --css "$(TEST3_HA_CSS)" --path test3.png

testrest:
	${BINDIR}/${HA_SS_NAME} --username "$(HA_USERNAME)" --password "$(HA_PASSWORD)" --restport 3500

testaviva: ${BINDIR}/${AVIVA_NAME}
	${BINDIR}/${AVIVA_NAME} --help
	${BINDIR}/${AVIVA_NAME} --username "$(AVIVA_USERNAME)" --password "$(AVIVA_PASSWORD)" --otpcleancommand "$(AVIVA_OTPCLEANCOMMAND)"  --otpcommand "$(AVIVA_OTPCOMMAND)"

testavivaselenium:
	OTP_CLEANCOMMAND="$(AVIVA_OTPCLEANCOMMAND)" OTP_COMMAND="$(AVIVA_OTPCOMMAND)" python3 ${AVIVA_NAME}/${AVIVA_NAME}.py

testavivamymoney: ${BINDIR}/${AVIVAMYMONEY_NAME}
	${BINDIR}/${AVIVAMYMONEY_NAME} --help
	${BINDIR}/${AVIVAMYMONEY_NAME} --username "$(AVIVAMYMONEY_USERNAME)" --password "$(AVIVAMYMONEY_PASSWORD)"  --word "$(AVIVAMYMONEY_WORD)"

testavivamymoneyselenium:
	python3 ${AVIVAMYMONEY_NAME}/${AVIVAMYMONEY_NAME}.py

testnutmeg: ${BINDIR}/${NUTMEG_NAME}
	${BINDIR}/${NUTMEG_NAME} --help
	${BINDIR}/${NUTMEG_NAME} --username "$(TEST1_NUTMEG_USERNAME)" --password "$(TEST1_NUTMEG_PASSWORD)" --otpcleancommand "$(NUTMEG_OTPCLEANCOMMAND)" --otpcommand "$(NUTMEG_OTPCOMMAND)"
	${BINDIR}/${NUTMEG_NAME} --username "$(TEST2_NUTMEG_USERNAME)" --password "$(TEST2_NUTMEG_PASSWORD)" --otpcleancommand "$(NUTMEG_OTPCLEANCOMMAND)" --otpcommand "$(NUTMEG_OTPCOMMAND)"

testfund: ${BINDIR}/${FUND_NAME}
	${BINDIR}/${FUND_NAME} --help
	${BINDIR}/${FUND_NAME} --fund "$(TEST1_FUND)"
	${BINDIR}/${FUND_NAME} --fund "$(TEST2_FUND)"
	${BINDIR}/${FUND_NAME} --fund "$(TEST3_FUND)"

testmoneyfarm: ${BINDIR}/${MONEYFARM_NAME}
	${BINDIR}/${MONEYFARM_NAME} --help
	${BINDIR}/${MONEYFARM_NAME} --username "$(TEST1_MONEYFARM_USERNAME)" --password "$(TEST1_MONEYFARM_PASSWORD)" --otpcleancommand "$(MONEYFARM_OTPCLEANCOMMAND)" --otpcommand "$(MONEYFARM_OTPCOMMAND)"
	${BINDIR}/${MONEYFARM_NAME} --username "$(TEST2_MONEYFARM_USERNAME)" --password "$(TEST2_MONEYFARM_PASSWORD)" --otpcleancommand "$(MONEYFARM_OTPCLEANCOMMAND)" --otpcommand "$(MONEYFARM_OTPCOMMAND)"

testmoneyhub: ${BINDIR}/${MONEYHUB_NAME}
	${BINDIR}/${MONEYHUB_NAME} --help
	MONEYHUB_BALANCE=$(shell ${BINDIR}/${NUTMEG_NAME} --username "$(TEST1_NUTMEG_USERNAME)" --password "$(TEST1_NUTMEG_PASSWORD)" --otpcleancommand "$(NUTMEG_OTPCLEANCOMMAND)" --otpcommand "$(NUTMEG_OTPCOMMAND)"); \
	${BINDIR}/${MONEYHUB_NAME} --username "$(TEST1_MONEYHUB_USERNAME)" --password "$(TEST1_MONEYHUB_PASSWORD)" --account "$(TEST1_MONEYHUB_ACCOUNT)" --account "$(TEST1_MONEYHUB_ACCOUNT)" --account "$(TEST1_MONEYHUB_ACCOUNT)" --account "$(TEST1_MONEYHUB_ACCOUNT)" --account "$(TEST1_MONEYHUB_ACCOUNT)" --balance $$MONEYHUB_BALANCE --balance $$MONEYHUB_BALANCE --balance $$MONEYHUB_BALANCE --balance $$MONEYHUB_BALANCE --balance $$MONEYHUB_BALANCE

testoctopuswheel: ${BINDIR}/${OCTOPUSWHEEL_NAME}
	${BINDIR}/${OCTOPUSWHEEL_NAME} --help
	${BINDIR}/${OCTOPUSWHEEL_NAME} --username "$(TEST1_OCTOPUS_USERNAME)" --password "$(TEST1_OCTOPUS_PASSWORD)"

testmycauseuk: ${BINDIR}/${MYCAUSEUK_NAME}
	${BINDIR}/${MYCAUSEUK_NAME} --help
	${BINDIR}/${MYCAUSEUK_NAME} --username "$(TEST1_MYCAUSEUK_USERNAME)" --password "$(TEST1_MYCAUSEUK_PASSWORD)"

clean:
	@go clean
	-@rm -rf ${BINDIR} ${LINUXARM64BINDIR} ${LINUXAMD64BINDIR} ${DARWINARM64BINDIR} ${WINDOWSAMD64BINDIR} test*.png *.zip 2>/dev/null || true