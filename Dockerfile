FROM golang AS build

RUN apt-get update && apt-get install ffmpeg -y && apt-get install opus-tools -y

WORKDIR /src
COPY src/ .
RUN go mod download

RUN go build -o go-discord-bot-tts .

ENTRYPOINT ["./go-discord-bot-tts", "-t", "YOUR_TOKEN"]