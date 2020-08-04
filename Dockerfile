FROM golang:1.13.4-alpine AS builder

RUN apk add bash g++ make wget --no-cache

WORKDIR /projects/nightfall_dlp

COPY Makefile go.mod go.sum ./
RUN make deps

# Install GolangCI-Lint
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.23.6
RUN wget -O - -q https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v0.9.15

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o nightfall_dlp ./cmd/nightfalldlp/

FROM alpine:3.8

RUN apk add bash --no-cache

WORKDIR /projects/nightfall_dlp

COPY --from=builder /projects/nightfall_dlp/nightfall_dlp .
COPY --from=builder /projects/nightfall_dlp/Makefile .

CMD ["./nightfall_dlp"]
