FROM golang:1.21.4-alpine

WORKDIR /craftyproxy

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /craftyreverseproxy

CMD [ "/craftyreverseproxy" ]