# SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
#
# SPDX-License-Identifier: CC0-1.0

FROM golang:1.25-bookworm AS base

FROM base AS build-env
WORKDIR /build
COPY lib/ /build/lib/
COPY transformers/sta-vm-to-gtfsrt/src/ /build/src/
WORKDIR /build/src
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/main

# BUILD published image
FROM alpine:3 AS build
WORKDIR /app
COPY --from=build-env /build/main .
ENTRYPOINT [ "./main"]

# LOCAL DEVELOPMENT
FROM base AS dev
WORKDIR /code
CMD ["go", "run", "."]

# TESTS
FROM base AS test
WORKDIR /code
CMD ["go", "test", "-v", "./..."]
