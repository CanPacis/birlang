build_win:
	windres -o main-res.syso res/bir.rc && go build -o bin/bir.exe -ldflags "-s -w" -i

build_linux:
	go build -o bin/bir -ldflags "-s -w" -i

run:
	cls && go run bir.go "C:/Users/tmwwd/go/src/birlang/test/test.bir"
	
repl:
	cls && go run bir.go