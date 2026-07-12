FROM docker.io/library/golang:1.26.5-alpine3.23 as builder

COPY . .

RUN go build

FROM docker.io/library/alpine:3.24

COPY --from=builder go/suslik /usr/local/bin

CMD ["suslik"]
