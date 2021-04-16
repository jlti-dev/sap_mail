FROM golang:1.16 as builder
WORKDIR /app
COPY go.mod go.mod
RUN go mod download
COPY app/ .
RUN GCO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/main .

FROM buildpack-deps:latest
WORKDIR /app
COPY start.sh /app/start.sh
COPY --from=builder /app/main /app/main
CMD ["/bin/bash", "/app/start.sh"]
#CMD /app/main
