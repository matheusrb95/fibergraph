FROM golang:1.24.5 as build
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -a -o correlation ./cmd/api/*.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=build /src/correlation .
EXPOSE 4000
CMD ["/correlation"]
