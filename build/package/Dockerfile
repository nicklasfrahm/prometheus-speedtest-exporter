FROM golang:1.20 AS build
ARG VERSION
ARG TARGET

WORKDIR /app
COPY go.* ./
RUN go mod download

ADD Makefile .
ADD cmd/ cmd/
RUN TARGET=$TARGET VERSION=$VERSION BINARY=exporter make build

FROM gcr.io/distroless/static:nonroot AS run
WORKDIR /
COPY --from=build /app/exporter .
USER 65532:65532
ENTRYPOINT [ "/exporter" ]
