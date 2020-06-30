FROM golang:1.13.3-stretch as builder
# Username and password to use basic auth and download the repo.
# Recommend using engbot for this
ARG GIT_USER
ARG GIT_PASS
ARG SERVICE_NAME

# Install GolangCI-Lint
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.23.6
RUN wget -O - -q https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v0.9.15

# verify the service name is provided
RUN test -n "$SERVICE_NAME"
WORKDIR /projects/$SERVICE_NAME
# Verify GIT_USER and GIT_PASS were passed in as arguments
RUN test -n "$GIT_USER"
RUN test -n "$GIT_PASS"
# Rewrite url to use basic auth from arguments passed in
RUN git config --global url."https://$GIT_USER:$GIT_PASS@github.com/".insteadOf "https://github.com/"
COPY Makefile go.mod go.sum ./
RUN make deps
COPY . .
RUN make generate
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $SERVICE_NAME .
