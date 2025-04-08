FROM docker.io/library/golang:1.24.1-alpine3.21 as builder

COPY . .

RUN go build

FROM docker.io/library/alpine:3.21

COPY --from=builder go/suslik /usr/local/bin

CMD ["suslik"]
