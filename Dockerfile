FROM alpine
COPY aws-delete-versioned-objects /
ENTRYPOINT ["/aws-delete-versioned-objects"]