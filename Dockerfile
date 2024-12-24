FROM golang:latest AS base

FROM base AS dev

RUN go install github.com/air-verse/air@latest
WORKDIR /opt/app/api

CMD ["air"]

# TODO: Add a production build