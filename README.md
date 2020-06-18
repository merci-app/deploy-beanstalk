# AWS Utils


### Build Docker Image
`$ docker build -t plugins/aws .`


#### Run

```bash
AWS_REGION=
AWS_ACCESS_KEY=
AWS_SECRET_KEY=
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
# AWS_DEGRADED_STATUS_TIMEOUT=15m # how long it will wait with status degraded
updateeb -application=name -environment=name -version=uniquie-hash -src=bucket:path
```
