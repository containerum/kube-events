FROM golang:1.10-alpine as builder
WORKDIR /go/src/github.com/containerum/kube-events
COPY . .
RUN go build -v -o /bin/kube-events ./cmd/kube-events

FROM alpine:3.7
COPY --from=builder /bin/kube-events /
ENV CONFIG="" \
  DEBUG="" \
  TEXT_LOG="" \
  MONGO_ADDRS="mongodb:27017" \
  MONGO_USER="user" \
  MONGO_PASSWORD="pass" \
  MONGO_DATABASE="kube-events" \
  MONGO_COLLECTION_SIZE="1073741824" \
  MONGO_COLLECTION_MAX_DOCS="" \
  BUFFER_CAPACITY="500" \
  BUFFER_FLUSH_PERIOD="30s"
CMD ["/kube-events"]
