# AWS Utils


### Build Docker Image
`$ docker build -t plugins/aws .`


#### Run

```bash
AWS_REGION=
AWS_ACCESS_KEY=
AWS_SECRET_KEY=

# optional=(default value)
MAX_RETRIES=20
MIN_RETRY_DELAY=1s
MAX_RETRY_DELAY=1m
```

```bash
# check if envs ENV ENV2 and ENV3 is already setted
checkenv -application=name -environment=name -envs="ENV ENV2 ENV3"

# check if a file already exists on S3
fileexistsons3 bucket:path/file.zip
echo $?
# not found exit with code 4

# upload file.zip to s3
uploads3 file.zip bucket:path/file.zip

# update Elastic Beanstalk Environment with S3 version
#
# Options
# AWS_CHECK_STATUS_INTERVAL=5s    # interval between status check
# AWS_CHECK_STATUS_TIMEOUT=15m    # how long it will wait until EB succeed or fail
# AWS_READY_STATUS_WAIT=15s       # how long it will wait with ready status
# AWS_DEGRADED_STATUS_TIMEOUT=15m # how long it will wait with status degraded
updateeb -application=name -environment=name -version=uniquie-hash -src=bucket:path
```
