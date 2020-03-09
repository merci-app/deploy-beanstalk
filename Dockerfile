FROM golang:1.14-alpine3.11
WORKDIR /go/src/deploy-beanstalk
ADD . .
RUN GOOS=linux CGO_ENABLED=0 go build -o bin/deploy

FROM alpine:3.11
RUN apk add --no-cache ca-certificates
COPY --from=0 /go/src/deploy-beanstalk/bin/deploy /bin/deploy-beanstalk
ENTRYPOINT ["/bin/deploy-beanstalk"]
