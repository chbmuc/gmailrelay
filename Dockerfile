FROM golang:latest AS build

WORKDIR /go/src/gmailrelay
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o gmailrelay .

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/src/gmailrelay/gmailrelay /gmailrelay

ENTRYPOINT ["/gmailrelay", "--config", "/etc/gmailrelay/gmailrelay.ini"]

