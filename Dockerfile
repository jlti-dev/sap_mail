FROM golang:1.17 as builder
WORKDIR /app
#COPY go.mod go.mod
COPY app/ .
run go mod init github.com/jlti-dev/sap_mail && go mod tidy
RUN GCO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/main .

FROM buildpack-deps:latest
RUN apt update && apt install -y iproute2
WORKDIR /app
COPY start.sh /app/start.sh
COPY --from=builder /app/main /app/main
CMD ["/bin/bash", "/app/start.sh"]
#CMD /app/main
