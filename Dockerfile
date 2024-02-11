FROM golang:1.21 AS builder
WORKDIR /go/src/fitsleepingishts
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/fitsleepingishts

# Copy only the binary in a new, lightweight container
FROM gcr.io/distroless/static-debian12
COPY --from=builder /go/bin/fitsleepingishts /

# Copy the website views and other stuff next to the binary
COPY --from=builder /go/src/fitsleepingishts/views/ /views/
COPY --from=builder /go/src/fitsleepingishts/static/ /static/

# Add useful stuff or stepping into the container and debug
COPY --from=busybox:1.36.1-uclibc /bin/sh /bin/sh
COPY --from=busybox:1.36.1-uclibc /bin/ls /bin/ls

CMD ["/fitsleepingishts"]
