FROM golang:latest AS builder
ADD . /source
WORKDIR /source
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /adwords-downloader .


FROM mesosphere/aws-cli:latest

RUN apk add --no-cache bash jq wget tar ca-certificates
RUN wget -O /usr/local/bin/yq "https://github.com/mikefarah/yq/releases/download/2.4.0/yq_linux_amd64"
RUN chmod +x /usr/local/bin/yq

COPY --from=builder /adwords-downloader ./
RUN chmod +x ./adwords-downloader

COPY platform-build/generate-local-config.sh ./generate-local-config.sh
RUN chmod +x ./generate-local-config.sh

COPY profiles ./profiles

COPY platform-build/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
