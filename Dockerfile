FROM golang

WORKDIR $GOPATH/src/github.com/telnyx

COPY . .

RUN go get -d -v ./...

RUN go install -v ./...

EXPOSE 8080

CMD ["telnyx"]