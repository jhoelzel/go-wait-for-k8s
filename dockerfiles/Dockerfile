FROM gcr.io/distroless/static:nonroot
WORKDIR /
ADD ./bin/go-wait-for-k8s /go-wait-for-k8s

USER 65532:65532
ENTRYPOINT ["/go-wait-for-k8s"]
