# Text to Speech Discord Bot

Discord bot for tts that autoplays after a `speak` command message is sent.  
It uses opus audio files and google tts api.

Written in go for fun and speed :)

## Running it

- Install [ffmpeg](https://ffmpeg.org/download.html) and [opus-tools](https://opus-codec.org/downloads/).
- Follow the instructions under [https://cloud.google.com/docs/authentication/production#passing_variable](https://cloud.google.com/docs/authentication/production#passing_variable)
  to generate Google TTS _(Text-to-Speech)_ API key file.
- Set the environment variable GOOGLE_APPLICATION_CREDENTIALS with the path to your key file.

```bash
go run main.go tts.go -t "YOUR_BOT_TOKEN"
```

or

```bash
go build -o go-discord-bot-tts .
./go-discord-bot-tts -t "YOUR_BOT_TOKEN"
```

### Using Docker _(recommended)_

- Edit the [Dockerfile](Dockerfile) to use your bot token.
- Follow the instructions under [https://cloud.google.com/docs/authentication/production#passing_variable](https://cloud.google.com/docs/authentication/production#passing_variable)
  to generate Google TTS _(Text-to-Speech)_ API key file.
- Paste the contents of your key in [gcloud-tts-api-key.json](gcloud-tts-api-key.json)

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
