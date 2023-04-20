FROM gcr.io/distroless/static:nonroot
WORKDIR /
ADD ./bin/go-wait-for /go-wait-for

USER 65532:65532
ENTRYPOINT ["/go-wait-for"]
