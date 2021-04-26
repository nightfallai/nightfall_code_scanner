FROM golang:1.16-alpine AS builder

RUN apk add bash g++ make wget --no-cache

WORKDIR /projects/nightfall_code_scanner

COPY Makefile go.mod go.sum ./
RUN make deps

# Install GolangCI-Lint
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.23.6
RUN wget -O - -q https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v0.9.15

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nightfall_code_scanner ./cmd/nightfalldlp/

FROM ubuntu:18.04

RUN apt update
RUN apt install -y git

COPY --from=builder /projects/nightfall_code_scanner/nightfall_code_scanner /nightfall_code_scanner

CMD ["/nightfall_code_scanner"]
