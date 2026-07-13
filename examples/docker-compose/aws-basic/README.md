# Docker Compose AWS subset example

Start Emulith and check health:

```bash
docker compose up --build
curl http://localhost:4566/_emulith/health
```

Use fake local credentials only:

```bash
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_REGION=us-east-1
export AWS_EC2_METADATA_DISABLED=true

aws --endpoint-url=http://localhost:4566 s3api create-bucket --bucket demo
aws --endpoint-url=http://localhost:4566 s3api put-object --bucket demo --key hello.txt --body README.md
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name demo
aws --endpoint-url=http://localhost:4566 sts get-caller-identity
```

The example is not validated against every AWS CLI version. Emulith implements only the documented POC subset.
