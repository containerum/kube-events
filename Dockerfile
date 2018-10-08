FROM golang:1.10-alpine as builder
WORKDIR /go/src/github.com/containerum/kube-events
COPY . .
RUN go build -v -o /bin/kube-events ./cmd/kube-events

FROM alpine:3.7
COPY --from=builder /bin/kube-events /
ENV CONFIG="" \
  DEBUG="" \
  TEXTLOG="" \
  MONGO_ADDR="mongodb:27017" \
  MONGO_LOGIN="user" \
  MONGO_PASSWORD="pass" \
  MONGO_DB="kube-events" \
  BUFFER_CAPACITY="500" \
  BUFFER_FLUSH_PERIOD="30s" \
  BUFFER_MIN_INSERT_EVENTS="1"
CMD ["/kube-events"]
