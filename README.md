# AWS Beanstalk Deploy


### Build
`$ docker build --rm=true -t plugins/deploy-beanstalk .`


### Run

```bash
# my.env
AWS_APPLICATION=               # application name
AWS_ENVIRONMENT=               # environment name in the app to update
AWS_REGION=                    # AWS region
AWS_ACCESS_KEY=                # AWS Access Key
AWS_SECRET_KEY=                # AWS Secret Key
AWS_BUCKET=                    # S3 bucket name
AWS_BUCKET_KEY=                # S3 file path
AWS_VERSION=                   # version label for the app(must be unique)
AWS_DESCRIPTION=               # description for the app version
AWS_UPLOAD=false               # upload file(AWS_BUCKT_KEY) to S3
AWS_AUTO_CREATE=false          # auto create app if it doesn't exist
AWS_CHECK_STATUS_INTERVAL=5s   # interval between status check
AWS_CHECK_STATUS_TIMEOUT=5m    # how long it will wait until EB succeed or fail
AWS_DEGRATE_STATUS_TIMEOUT=25s # how long it will wait with status degraded

# run with go
$ go install
$ export $(xargs < my.env)
$ deploy-beanstalk

# run with docker
$ docker run --rm --env-file ./my.env -v ${PWD}:/app -w /app plugins/deploy-beanstalk
```
