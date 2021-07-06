build:
	go build ./...
	go install ./...
run:
	gomud
debug:
	MUDDEBUG=1 GODEBUG=memprofilerate=1 gomud