FROM golang:1.14
WORKDIR /go/src/deploy-beanstalk
ADD . .
RUN GOOS=linux CGO_ENABLED=0 go build -o bin/deploy cmd/deploy/deploy.go
RUN GOOS=linux CGO_ENABLED=0 go build -o bin/checkenv cmd/checkenv/checkenv.go
RUN GOOS=linux CGO_ENABLED=0 go build -o bin/uploads3 cmd/uploads3/uploads3.go
RUN GOOS=linux CGO_ENABLED=0 go build -o bin/fileexistsons3 cmd/fileexistsons3/fileexistsons3.go
RUN GOOS=linux CGO_ENABLED=0 go build -o bin/updateeb cmd/updateeb/updateeb.go
RUN apt update && apt install zip -y

FROM golang:1.14
COPY --from=0 /go/src/deploy-beanstalk/bin/deploy /bin/deploy-beanstalk
COPY --from=0 /go/src/deploy-beanstalk/bin/* /usr/local/bin/
COPY --from=0 /usr/bin/zip /usr/bin/
COPY --from=0 /usr/bin/unzip /usr/bin/
ENTRYPOINT ["deploy-beanstalk"]
