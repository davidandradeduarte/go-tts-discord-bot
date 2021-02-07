FROM golang AS build

RUN apt-get update && apt-get install ffmpeg -y && apt-get install opus-tools -y
RUN mkdir /usr/gcloud

COPY gcloud-tts-api-key.json /usr/.gcloud/
ENV GOOGLE_APPLICATION_CREDENTIALS=/usr/.gcloud/gcloud-tts-api-key.json

WORKDIR /src
COPY src/ .
RUN go mod download

RUN go build -o go-discord-bot-tts .

ENTRYPOINT ["./go-discord-bot-tts", "-t", "YOUR_BOT_TOKEN"]