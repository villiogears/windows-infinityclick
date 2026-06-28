include .env
export

build:
	go build -o test.exe .

clean:
	rm -f test.exe
	rm -rf fyne-cross

fyne-cross:
	fyne-cross windows -arch=amd64 -app-id com.elphadeal.test -icon icon.png --output test.exe

fyne-cross-mac:
	fyne-cross darwin -arch=arm64 -app-id com.elphadeal.test -icon icon.png --output test.app