# Name of the binary output
BINARY = veloherodown

# Values to pass as VERSION
# GitHub Tag name: https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
VERSION = ${GITHUB_REF_NAME}
ifeq ($(VERSION),)
	VERSION = `git rev-parse --short HEAD`
endif

# Builds the project
build: native linux macos windows

native:
	go build -o ${BINARY}
	@echo

linux:
	@echo "Linux"
	GOOS=linux GOARCH=amd64 go build -o ${BINARY}-linux-x86_64
	GOOS=linux GOARCH=arm64 go build -o ${BINARY}-linux-arm64
	@echo

macos:
	@echo "macOS"
	GOOS=darwin GOARCH=amd64 go build -o ${BINARY}-macos-x86_64
	GOOS=darwin GOARCH=arm64 go build -o ${BINARY}-macos-arm64
	@echo

windows:
	@echo "Windows"
	GOOS=windows GOARCH=amd64 go build -o ${BINARY}-windows-x86_64.exe
	GOOS=windows GOARCH=arm64 go build -o ${BINARY}-windows-arm64.exe
	@echo

# Clean the project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	if [ -f ${BINARY}-linux-x86_64 ] ; then rm ${BINARY}-linux-x86_64 ; fi
	if [ -f ${BINARY}-linux-arm64 ] ; then rm ${BINARY}-linux-arm64 ; fi
	if [ -f ${BINARY}-macos-x86_64 ] ; then rm ${BINARY}-macos-x86_64 ; fi
	if [ -f ${BINARY}-macos-arm64 ] ; then rm ${BINARY}-macos-arm64 ; fi
	if [ -f ${BINARY}-windows-x86_64.exe ] ; then rm ${BINARY}-windows-x86_64.exe ; fi
	if [ -f ${BINARY}-windows-arm64.exe ] ; then rm ${BINARY}-windows-arm64.exe ; fi
