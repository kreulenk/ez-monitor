build:
	@go build -v -o bin/ez-monitor

install: build
	 @cp ./bin/ez-monitor /usr/local/bin/