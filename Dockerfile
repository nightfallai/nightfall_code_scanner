FROM golang:1.13.3-stretch as builder
# Username and password to use basic auth and download the repo.
# Recommend using engbot for this
ARG GIT_USER
ARG GIT_PASS
ARG SERVICE_NAME
ARG GRPC_HEALTH_PROBE_VERSION=v0.3.2

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

FROM alpine:3.8
# ARG's aren't global to the Dockerfile, they're per stage so we need
# to redeclare this here to be used below in build commands like RUN & WORKDIR
ARG SERVICE_NAME
# This env variable needs to be set to access the service name at runtime,
# primarily for the CMD that runs on startup
ENV SERVICE_NAME=$SERVICE_NAME
# These lines are for establishing datadog checks. go_expvar uses the standard
# /debug/vars route to query for metrics from this service. http_check uses
# the /ping route and watches for 200 status codes. The special host and port
# syntax is from datadogs autodiscovery feature. The port syntax finds the
# highest exposed port sorted numerically.
LABEL "com.datadoghq.ad.check_names"='["go_expvar", "http_check"]'
LABEL "com.datadoghq.ad.init_configs"='[{}, {}]'
LABEL "com.datadoghq.ad.instances"='[\
  { "expvar_url": "http://%%host%%:%%port%%" },\
  {\
    "name": "%%env_SERVICE_NAME%%",\
    "url": "http://%%host%%:%%port%%/ping",\
    "http_response_status_code": 200\
  }\
]'
WORKDIR /projects/$SERVICE_NAME/
RUN mkdir -p /var/log/watchtower/$SERVICE_NAME/
RUN apk add ca-certificates
COPY --from=builder /projects/$SERVICE_NAME/ .
ENV GO_ENV=production
ENV GIN_MODE=release
EXPOSE 8080
CMD ./${SERVICE_NAME}
