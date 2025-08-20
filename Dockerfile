# Doing a multi-stage build to make sure to have passing of unittests enforced
FROM docker.io/golang:1.24 AS base

LABEL org.opencontainers.image.source=https://github.com/pvbouwel/run-pod-spec
LABEL org.opencontainers.image.description="Run a Kubernetes pod manifest, stream the logs and cleanup the pod at the end."
LABEL org.opencontainers.image.licenses=AGPL-3.0

COPY go.mod /usr/src/runpodspec/go.mod
COPY go.sum /usr/src/runpodspec/go.sum
WORKDIR /usr/src/runpodspec
# To not have to fetch dependencies each build
RUN go mod download
RUN go mod tidy
ADD pkg /usr/src/runpodspec/pkg
ADD main.go /usr/src/runpodspec/
RUN go get
RUN go test -p 1 -coverprofile cover.out ./...
RUN go vet
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# If tests pass then build the final stage
FROM scratch

# We need to trust common SSL certificates to get issuer info
COPY --from=base /etc/ssl /etc/ssl
# We need our binary
COPY --from=base /usr/src/runpodspec/main /run-pods-pec

ENTRYPOINT [ "/run-pods-pec" ]
