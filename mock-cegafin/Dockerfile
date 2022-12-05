FROM golang:1.16-alpine
WORKDIR /source

COPY ./mock-cegafin/go.mod ./
COPY ./mock-cegafin/go.sum ./

RUN go mod download

COPY ./mock-cegafin/mockcega.go ./

RUN CGO_ENABLED=0 go build -o /bin/mock-cegafin

CMD [ "/bin/mock-cegafin" ]
