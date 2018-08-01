FROM golang:1.9.2-stretch
RUN mkdir -p /go/src/github.com/tezexInfo/TezProxy
ADD . /go/src/github.com/tezexInfo/TezProxy
WORKDIR /go/src/github.com/tezexInfo/TezProxy
RUN go get -d .
RUN CGO_ENABLED=0 go build -o build/app .

FROM scratch
WORKDIR /
COPY --from=0 /go/src/github.com/tezexInfo/TezProxy/build/app /app
CMD ["/app"]