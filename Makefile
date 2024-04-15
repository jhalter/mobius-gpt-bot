linux_bot_arm_target=dist/mobius_bot_linux_arm
build-linux-arm-bot:
	mkdir $(linux_bot_arm_target) ; true
	GOOS=linux GOARCH=arm go build -o $(linux_bot_arm_target)/mobius-hotline-bot main.go

package-linux-arm-bot: build-linux-arm-bot
	tar -zcvf $(linux_bot_arm_target).tar.gz $(linux_bot_arm_target)

linux_bot_amd64_target=dist/mobius_bot_linux_amd64
build-linux-amd64-bot:
	mkdir $(linux_bot_amd64_target) ; true
	GOOS=linux GOARCH=amd64 go build -o $(linux_bot_amd64_target)/mobius-hotline-bot main.go

package-linux-amd64-bot: build-linux-amd64-bot
	tar -zcvf $(linux_bot_amd64_target).tar.gz $(linux_bot_amd64_target)

darwin_bot_amd64_target=dist/mobius_bot_darwin_amd64
build-darwin-amd64-bot:
	mkdir $(darwin_bot_amd64_target) ; true
	GOOS=darwin GOARCH=amd64 go build -o $(darwin_bot_amd64_target)/mobius-hotline-bot main.go

package-darwin-amd64-bot: build-darwin-amd64-bot
	tar -zcvf dist/mobius_bot_darwin_amd64.tar.gz $(darwin_bot_amd64_target)

windows_bot_amd64_target=dist/mobius_bot_windows_amd64
build-win-amd64-bot:
	mkdir $(windows_bot_amd64_target) ; true
	GOOS=windows GOARCH=amd64 go build -o $(windows_bot_amd64_target)/mobius-hotline-bot.exe main.go

package-win-amd64-bot: build-win-amd64-bot
	zip -r dist/mobius_bot_windows_amd64.zip $(windows_bot_amd64_target)

all: clean \
	package-win-amd64-bot \
 	package-darwin-amd64-bot \
 	package-linux-arm-bot \
 	package-linux-amd64-bot \

clean:
	rm -rf dist/*