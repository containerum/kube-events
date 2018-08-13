FROM golang:1.10-alpine as builder
WORKDIR /go/src/github.com/containeurm/kube-events
COPY . .
RUN go build -v -o /bin/kube-events ./cmd/kube-events

FROM alpine:3.7
COPY --from=builder /bin/kube-events /
ENV CONFIG="" \
  NAMESPACE="" \
  LABEL_SELECTOR="" \
  FIELD_SELECTOR=""
CMD ["/kube-events"]
