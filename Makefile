all: run

run:
	(cd src && go run main.go tts.go -t "YOUR_BOT_TOKEN")