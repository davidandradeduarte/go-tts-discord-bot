# Text to Speech Discord Bot

![Pipeline](https://github.com/davidandradeduarte/tts-discord-bot/actions/workflows/pipeline.yml/badge.svg)

Discord bot for tts that autoplays after a `speak` command message is sent.  
It uses opus audio files and google tts api.

Written in go for fun and speed :)

## Running it

- Follow the instructions under [https://cloud.google.com/docs/authentication/production#passing_variable](https://cloud.google.com/docs/authentication/production#passing_variable)
  to generate a Google TTS _(Text-to-Speech)_ API key file

### With Docker _(recommended)_

- Edit the [Dockerfile](Dockerfile) to use your bot token
- Paste the contents of your key in [gcloud-tts-api-key.json](gcloud-tts-api-key.json)

```bash
make docker
# or
docker build -t tts-discord-bot -f Dockerfile --no-cache .
docker run tts-discord-bot
```

### Without Docker

- Install [ffmpeg](https://ffmpeg.org/download.html)
- Install [opus-tools](https://opus-codec.org/downloads/)
- Set the environment variable GOOGLE_APPLICATION_CREDENTIALS with the path to your key file

```bash
make run
# or
go run main.go tts.go -t "YOUR_BOT_TOKEN"
```

## Usage

Type: `speak` `<Your Message>` in a discord chat.  
The bot will read the message, call the TTS API, convert mp3 to opus, connect
to your current voice channel and will play your message.
After it's done playing, disconnects.

Currently, only Portuguese is supported.
