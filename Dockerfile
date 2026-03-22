# Stage 1: Build UI
FROM oven/bun:1-alpine AS ui-builder
WORKDIR /app/ui
COPY ui/package.json ui/bun.lock ./
RUN bun install --frozen-lockfile
COPY ui/ .
RUN bun run build

# Stage 2: Build Go binary
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/ui/build ./ui/build
RUN go build -o picowatch .

# Stage 3: Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/picowatch .
EXPOSE 8080
CMD ["./picowatch"]
