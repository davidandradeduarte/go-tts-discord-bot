# Text to Speech Discord Bot

Discord bot for tts that auto plays after a `speak` command message is sent.  
It uses opus audio files and a free tts php api (will migrate to google later).

Written in go for fun and the speed :)

## Running it

Edit [Dockerfile](Dockerfile) to use your bot token.  

```bash
docker build -t local/go-tts-discord-bot -f Dockerfile --no-cache .
docker run local/go-tts-discord-bot
```

## Usage

Type: `speak` `<Your Message>` in a discord chat.  
The bot will read the message, call the TTS API, convert mp3 to opus, connect
to your current voice channel and will play your message.
After it's done playing, disconnects.

Currently, only Portuguese is supported.

## TODO
- [x] Add dependencies and go modules
- [x] Create a Dockerfile to run inside a container
- [ ] Add Dockerfile env variable for bot token
- [ ] Migrate to Google TTS api.
- [ ] Stay connected during X minutes. Disconnect if it doesn't receive any more requests.
- [ ] Add support for multiple languages