FROM golang:1.12.0-alpine3.9 AS builder

RUN apk add git
ADD . /src
RUN cd /src && go build -o app


FROM alpine
WORKDIR /app
COPY --from=builder /src/app /app/
CMD ["./app"]
