FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY gh-bootstrap-repository gh-bootstrap-repository
USER 65532:65532

ENTRYPOINT ["/gh-bootstrap-repository"]
