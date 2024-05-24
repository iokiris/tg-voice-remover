FROM golang:1.22.1 as development

WORKDIR /app

RUN apt-get update && apt-get install -y jq

RUN go install github.com/githubnemo/CompileDaemon@latest

COPY go.mod ./
COPY go.sum ./
RUN go mod tidy

COPY . .

LABEL org.opencontainers.image.source https://github.com/iokiris/tg-voice-remover


CMD ["CompileDaemon", "-polling=true", "-log-prefix=true", "-build=go build -o ./tmp/myapp ./", "-command=./tmp/myapp"]
