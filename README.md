# Text to Speech Discord Bot

Discord bot for tts that auto plays after a `speak` command message is sent.  
It uses opus audio files and a free tts php api (will migrate to google later).  

Written in go for fun and the speed :)  

It's still a WIP but it works.

## Running it

```bash
docker build -t local/go-tts-discord-bot -f Dockerfile --no-cache .
docker run local/go-tts-discord-bot --env BOT_TOKEN=YOUR_TOKEN
```
## TODO
- [x] Add dependencies and go modules
- [x] Create a Dockerfile to run inside a container
- [ ] Migrate to Google TTS api.
- [ ] Stay connected during X minutes. Disconnect if it doesn't receive any more requests.