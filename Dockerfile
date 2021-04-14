FROM golang:latest as builder
WORKDIR /app
RUN go get github.com/xhit/go-simple-mail && \
	go get github.com/prometheus/client_golang/prometheus && \
	go get github.com/prometheus/client_golang/prometheus/promhttp
COPY app/ .
RUN GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main
RUN rm -rf /app/*.go

FROM buildpack-deps:stable
WORKDIR /app
COPY start.sh /app/start.sh
COPY --from=builder /app /app
CMD ["/bin/bash", "/app/start.sh"]
