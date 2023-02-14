FROM golang:1.19 as builder
WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make

FROM alpine
WORKDIR /
COPY --from=builder /workspace/s3proxy .
COPY --from=builder /workspace/encrypt .
COPY --from=builder /workspace/decrypt .
ENTRYPOINT ["/s3proxy"]
