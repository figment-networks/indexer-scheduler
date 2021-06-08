# ------------------------------------------------------------------------------
# Node Builder Image
# ------------------------------------------------------------------------------
FROM node AS build-ui

WORKDIR /build

COPY ./Makefile ./Makefile
COPY ./assets ./assets


RUN make prepare-ui-install-modules
RUN make prepare-ui

# ------------------------------------------------------------------------------
# Node Builder Image
# ------------------------------------------------------------------------------
FROM golang AS build

WORKDIR /build

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download

COPY ./Makefile ./Makefile
COPY ./cmd ./cmd
COPY ./conn ./conn
COPY ./core ./core
COPY ./destination ./destination
COPY ./persistence ./persistence
COPY ./process ./process
COPY ./runner ./runner
COPY ./structures ./structures
COPY ./ui/ui.go ./ui/ui.go



COPY --from=build-ui  /build/ui /build/ui

ENV CGO_ENABLED=0
ENV GOARCH=amd64
ENV GOOS=linux

RUN \
  GO_VERSION=$(go version | awk {'print $3'}) \
  GIT_COMMIT=$(git rev-parse HEAD) \
  make build && make build-migration

# ------------------------------------------------------------------------------
# Target Image
# ------------------------------------------------------------------------------
FROM alpine AS release

WORKDIR /app/

COPY --from=build /build/migration /app/migration
COPY ./cmd/scheduler-migration/migrations/ /app/migrations/
COPY --from=build  /build/scheduler /app/scheduler

CMD ["/app/scheduler"]
