package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"k8s.io/utils/pointer"
)

var (
	name         = "athena-query-executor"
	version      = "dev"
	commit       string
	athenaClient *athena.Athena
	s3Downloader *s3manager.Downloader
)

const (
	waitDuration       = 5 * time.Second
	fileNameDateFormat = "20060102150405"
)

func init() {
	cred := credentials.NewStaticCredentials(
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		os.Getenv("AWS_SESSION_TOKEN"),
	)
	conf := aws.Config{
		Region:      aws.String(os.Getenv("AWS_DEFAULT_REGION")),
		Credentials: cred,
	}
	sess, err := session.NewSession(&conf)

	if err != nil {
		log.Fatal(err)
	}

	athenaClient = athena.New(sess)
	s3Downloader = s3manager.NewDownloader(sess)
}

func main() {
	showVersion := false
	query := flag.String("query", "", "please specify -query flag")
	saveBucket := flag.String("result-save-bucket", "", "please specify -result-save-bucket flag")

	flag.BoolVar(&showVersion, "version", false, "show application version")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s version:%s (rev:%s)\n", name, version, commit)
		os.Exit(0)
	}

	resultConf := &athena.ResultConfiguration{}

	if *saveBucket == "" {
		if os.Getenv("ATHENA_RESULT_BUCKET") == "" {
			log.Fatal("Please specify S3 bucket for saving Athena query results.")
		}

		saveBucket = pointer.StringPtr(os.Getenv("ATHENA_RESULT_BUCKET"))
	}

	log.Printf("query: %s", *query)
	log.Printf("saveBucket: %s", *saveBucket)

	resultConf.SetOutputLocation("s3://" + *saveBucket + "/")

	input := &athena.StartQueryExecutionInput{
		QueryString:         query,
		ResultConfiguration: resultConf,
	}

	log.Printf("Execute query: %s", *query)

	executionResult, err := getQueryExecutionResultID(input)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Query Succeed. S3Output path: %s", *executionResult.QueryExecution.ResultConfiguration.OutputLocation)

	u, err := url.Parse(*executionResult.QueryExecution.ResultConfiguration.OutputLocation)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("proto: %q, bucket: %q, key: %q", u.Scheme, u.Host, u.Path)

	f, err := os.Create(fmt.Sprintf("%s%s", time.Now().Format(fileNameDateFormat), filepath.Ext(u.Path)))
	if err != nil {
		log.Fatal(err)
	}

	n, err := s3Downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(u.Path),
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("FileName: %s Size: %d byte", f.Name(), n)
}

func getQueryExecutionResultID(
	input *athena.StartQueryExecutionInput) (
	executionOutput *athena.GetQueryExecutionOutput, err error) {
	output, err := athenaClient.StartQueryExecution(input)

	if err != nil {
		return nil, err
	}

	id := output.QueryExecutionId
	executionInput := &athena.GetQueryExecutionInput{
		QueryExecutionId: id,
	}

	for {
		executionOutput, err = athenaClient.GetQueryExecution(executionInput)
		if err != nil {
			return nil, err
		}
		// @see https://docs.aws.amazon.com/sdk-for-go/api/service/athena/#pkg-consts

		switch *executionOutput.QueryExecution.Status.State {
		case athena.QueryExecutionStateQueued, athena.QueryExecutionStateRunning:
			time.Sleep(waitDuration)
		case athena.QueryExecutionStateSucceeded:
			return executionOutput, nil
		default: // athena.QueryExecutionStateFailed, athena.QueryExecutionStateCancelled
			return nil, fmt.Errorf("error: %v", executionOutput.String())
		}
	}
}
