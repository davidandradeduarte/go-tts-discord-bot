run:
	(cd src && go run main.go tts.go -t "YOUR_BOT_TOKEN")

docker:
	docker build -t tts-discord-bot -f Dockerfile .
	docker run -d --name tts-discord-bot tts-discord-bot