ARG GOVER=1.20.1
ARG BASE_DISTRO=alpine
ARG GITID
FROM golang:${GOVER}-${BASE_DISTRO} AS build

# Steps to build
RUN apk add --no-cache shadow=4.13-r0 && \
        useradd -u 10000 -ms /bin/ash build-user

COPY --chown=10000:10000 go.* /home/build-user/service/

USER 10000

RUN mkdir -p /home/build-user/service/cloud-builder

WORKDIR /home/build-user/service

RUN go mod download

USER root

COPY --chown=10000:10000 . /home/build-user/service

COPY --chown=10000:10000 ./cloud-builder /home/build-user/service/cloud-builder

USER 10000

RUN CGO_ENABLED=0 go build -ldflags '-w' -o /home/build-user/service/out/service ./server

# hadolint ignore=DL3007
FROM gcr.io/distroless/static:latest AS release
ARG GOVER
ARG GITID
LABEL   org.opencontainers.image.created="2021-08-03" \
        org.opencontainers.image.version="0.0.1-alpha" \
        org.opencontainers.image.revision=$GITID \
        org.opencontainers.image.title="Decoupled Provision Service" \
        io.pantheon.golang-version=$GOVER

COPY --from=build /home/build-user/service/out/service /

# Copy required cloud-builder folder.
COPY --from=build /home/build-user/service/cloud-builder/*.yaml /assets/

ENV     PORT=8080 \
        RCB_PORT=8080 \
        RCB_HOST=0.0.0.0

EXPOSE 8080

ENTRYPOINT [ "/service" ]