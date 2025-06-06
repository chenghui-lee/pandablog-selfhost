# Start by building the application.
FROM golang:1.23 as build

WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN touch .env && make test

RUN CGO_ENABLED=0 go build -trimpath -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/static
COPY --from=build /go/bin/app /
CMD ["/app"]
