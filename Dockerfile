FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/cybersec ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/cybersec /app/cybersec
COPY templates /app/templates
COPY static /app/static
VOLUME ["/app/data", "/app/logs"]
EXPOSE 5000
USER nonroot:nonroot
ENTRYPOINT ["/app/cybersec"]
