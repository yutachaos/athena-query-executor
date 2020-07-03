# Athena query executor

- Executes a query on Athena and gets the result and download csv locally

## Require
- Settings are read from environment variables, so AWS authentication key is required

```sh
AWS_ACCESS_KEY_ID=xxxx
AWS_SECRET_ACCESS_KEY=xxxx
AWS_DEFAULT_REGION=xxx
// Specify with S3 path to save the execution result of Athena, instead of -result-save-bucket flag
ATHENA_RESULT_BUCKET=S3Path
```

## Usage
- query
     - The query to execute
- result-save-bucket
     - Bucket where the query execution result is saved

### go run
- go run main.go -query "SHOW DATABASES" -result-save-bucket hoge-bucket


