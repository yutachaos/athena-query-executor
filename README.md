# Athena query executor

- Executes a query on Athena and gets the result and download csv locally

## Require
- Settings are read from environment variables, so AWS authentication key is required

```sh
AWS_ACCESS_KEY_ID=xxxx
AWS_SECRET_ACCESS_KEY=xxxx
AWS_DEFAULT_REGION=xxx
// Specify with S3 path to save the execution result of Athena, instead of -save-bucket flag
AWS_S3_BUCKET_FOR_ATHENA_RESULT=S3Path
```

## Usage
- query
     - The query to execute
- save-bucket
     - Bucket where the query execution result is saved

### go run
- go run main.go -query "SHOW DATABASES" -save-bucket hoge-bucket


