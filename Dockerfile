FROM golang:1.13.3-stretch as builder

ARG SERVICE_NAME

# Install GolangCI-Lint
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.23.6
RUN wget -O - -q https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v0.9.15

# verify the service name is provided
RUN test -n "$SERVICE_NAME"
WORKDIR /projects/$SERVICE_NAME
COPY Makefile go.mod go.sum ./
RUN make deps
COPY . .
RUN make generate
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $SERVICE_NAME .
