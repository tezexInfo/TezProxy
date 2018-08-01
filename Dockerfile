FROM golang:1.9.2-stretch
RUN mkdir -p /go/src/gitlab.com/bitfly/TezProxy/build
ADD . /go/src/gitlab.com/bitfly/TezProxy
WORKDIR /go/src/gitlab.com/bitfly/TezProxy
RUN go get -d .
RUN CGO_ENABLED=0 go build -o build/app .

FROM scratch
WORKDIR /
COPY --from=0 /go/src/gitlab.com/bitfly/TezProxy/build/app /app
CMD ["/app"]