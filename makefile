build:
	windres -o main-res.syso res/bir.rc && go build -o bin/bir.exe -ldflags "-w" -i

run:
	cls && go run bir.go "C:/Users/tmwwd/go/src/bir/test/test.bir"
	
repl:
	cls && go run bir.go