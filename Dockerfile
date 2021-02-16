FROM golang:latest AS builder
ADD . /source
WORKDIR /source
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /adwords-downloader .


FROM mesosphere/aws-cli:latest

COPY --from=builder /adwords-downloader ./
RUN chmod +x ./adwords-downloader

ENTRYPOINT ["./adwords-downloader"]
