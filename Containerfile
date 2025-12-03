FROM docker.io/library/golang:1.25.5-alpine3.21 as builder

COPY . .

RUN go build

FROM docker.io/library/alpine:3.23

COPY --from=builder go/suslik /usr/local/bin

CMD ["suslik"]
