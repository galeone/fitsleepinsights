FROM golang:1.21 AS builder
WORKDIR /go/src/fitsleepingishts
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/fitsleepingishts

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/fitsleepingishts /
CMD ["/fitsleepingishts"]
