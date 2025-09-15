# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache make
RUN make build-goose
RUN make linux

# Final image
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/fury-flow-linux ./fury-flow
COPY --from=builder /app/bin/goose ./goose
COPY migrations ./migrations
COPY entrypoint.sh ./
RUN chmod +x entrypoint.sh
EXPOSE 8080
ENTRYPOINT [ "./entrypoint.sh" ]