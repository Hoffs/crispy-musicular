FROM golang:1.16-buster as build

WORKDIR /go/src/appbuild

ADD . .

RUN mkdir out
RUN go build -o ./out/app -v ./cmd/crispy_musicular

FROM golang:1.16-buster as app

WORKDIR /go/src/app

COPY --from=build /go/src/appbuild/out/app .
COPY --from=build /go/src/appbuild/templates /go/src/app/templates

CMD ["./app"]
