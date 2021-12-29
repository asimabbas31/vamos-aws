package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
)

func getvalue() {
	var access, secret, region string
	fmt.Println("ENTER YOUR ACCESS KEY")
	fmt.Scanln(&access)
	fmt.Println("ENTER YOUR SECRET KEY")
	fmt.Scanln(&secret)
	fmt.Println("ENTER YOUR REGION")
	fmt.Scanln(&region)
	os.Setenv("AWS_ACCESS_KEY", access)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secret)
	os.Setenv("AWS_REGION", region)
	os.Setenv("AWS_DEFAULT_OUTPUT", "json")
}

func awssess() *session.Session {
	var mfaCode string
	_iam := iam.New(session.New())
	devices, err := _iam.ListMFADevices(&iam.ListMFADevicesInput{})
	sn := devices.MFADevices[0].SerialNumber
	if err != nil {
		panic(err)
	}
	fmt.Printf("%1s\n", *sn)

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := sts.New(sess)
	fmt.Println("ENTER Token")
	fmt.Scanln(&mfaCode)

	params := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(900),
		SerialNumber:    aws.String(*sn),
		TokenCode:       aws.String(mfaCode),
	}
	resp, err := svc.GetSessionToken(params)

	fmt.Println(awsutil.StringValue(resp.Credentials.SessionToken))
	os.Setenv("AWS_SESSION_TOKEN", awsutil.StringValue(resp.Credentials.SessionToken))
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return sess
	}

	return sess

	// Pretty-print the response data.
}

func list() {
	svc := s3.New(session.New())

	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func ssid(sess *session.Session) {

	ssmsvc := ssm.New(sess)
	param, err := ssmsvc.GetParametersByPath(&ssm.GetParametersByPathInput{
		Path:           aws.String("/dev/"),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		panic(err)
	}

	value := param.Parameters[0:1]
	fmt.Println(value)
}

func main() {
	var sess *session.Session
	getvalue()
	sess = awssess()
	ssid(sess)
}
