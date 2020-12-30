FROM golang:1.15.6 as builder

WORKDIR /go/src/github.com/alexlast/ecr-credential-updater

COPY go.mod .
COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -a -o updater github.com/alexlast/ecr-credential-updater/cmd/updater

FROM alpine:3.12

WORKDIR /opt/updater

COPY --from=builder /go/src/github.com/alexlast/ecr-credential-updater/updater .

ENTRYPOINT ["./updater"]
