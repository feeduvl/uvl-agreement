FROM golang:1.15
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 9662
CMD ["app"]
