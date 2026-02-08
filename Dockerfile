# Build Stage
FROM golang:1.21-alpine AS builder

# Arbeitsverzeichnis setzen
WORKDIR /app

# CGO deaktivieren für statisches Binary
ENV CGO_ENABLED=0

# Installiere git (erforderlich für `go mod download` bei direkten VCS-Abhängigkeiten)
RUN apk add --no-cache git openssh

# Go Module herunterladen
# Kopiere go.mod und go.sum, lade Module ohne Verify (SumDB deaktiviert)
COPY go.mod go.sum ./
ENV GOSUMDB=off
RUN go env -w GOPROXY=direct && \
    go mod download

# Quellcode kopieren
COPY . .

# Binary bauen
RUN go build -ldflags="-s -w" -o /temp-recorder ./cmd/temp-recorder

# Runtime Stage
FROM alpine:3.19

# Labels für OCI-Kompatibilität
LABEL org.opencontainers.image.title="Temperature Recorder"
LABEL org.opencontainers.image.description="Records temperature data from serial port to MySQL database"
LABEL org.opencontainers.image.source="https://github.com/your-repo/temp-recorder"

# Notwendige Pakete für serielle Schnittstelle
RUN apk --no-cache add ca-certificates tzdata

# Nicht-root Benutzer erstellen (mit Device-Zugriff)
RUN addgroup -g 20 -S dialout || true && \
    adduser -S -u 1000 -G dialout appuser

# Binary kopieren
COPY --from=builder /temp-recorder /usr/local/bin/temp-recorder

# Arbeitsverzeichnis setzen
WORKDIR /app

# Benutzer wechseln
USER appuser

# Healthcheck (optional)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep temp-recorder || exit 1

# Anwendung starten
ENTRYPOINT ["/usr/local/bin/temp-recorder"]
