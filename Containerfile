FROM docker.io/library/golang:1.23.3-alpine3.20 as builder

COPY . .

RUN go build

FROM docker.io/library/alpine:3.20

COPY --from=builder go/suslik /usr/local/bin

CMD ["suslik"]
