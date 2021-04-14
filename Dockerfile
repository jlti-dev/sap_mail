FROM golang:1.16 as builder
WORKDIR /app
COPY go.mod go.mod
RUN go mod download
COPY app/ .
RUN GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main
RUN rm -rf /app/*.go

FROM buildpack-deps:stable
WORKDIR /app
COPY start.sh /app/start.sh
COPY --from=builder /app /app
CMD ["/bin/bash", "/app/start.sh"]
