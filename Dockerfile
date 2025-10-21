FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/lightboard-vmix-bridge

FROM scratch
COPY --from=builder /app/lightboard-vmix-bridge /lightboard-vmix-bridge

ENTRYPOINT [ "/lightboard-vmix-bridge" ]