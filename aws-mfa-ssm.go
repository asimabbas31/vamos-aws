package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
	"golang.org/x/crypto/ssh/terminal"
)

// To get the User input for Session
func getvalue() {
	var region string
	fmt.Println("***ENTER YOUR ACCESS KEY***")
	access, err := terminal.ReadPassword(0)
	fmt.Println(access)
	fmt.Println("***ENTER YOUR SECRET KEY***")
	secret, err := terminal.ReadPassword(0)
	fmt.Println(secret)
	fmt.Println("ENTER YOUR REGION")
	fmt.Scanln(&region)
	os.Setenv("AWS_ACCESS_KEY", string(access))
	os.Setenv("AWS_SECRET_ACCESS_KEY", string(secret))
	os.Setenv("AWS_REGION", region)
	os.Setenv("AWS_DEFAULT_OUTPUT", "json")

	if err != nil {
	}

}

// Create a Session using input values and get the Session Token
func awssess() *session.Session {
	var mfaCode string
	_iam := iam.New(session.New())
	devices, err := _iam.ListMFADevices(&iam.ListMFADevicesInput{})
	sn := devices.MFADevices[0].SerialNumber
	if err != nil {
		panic(err)
	}

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := sts.New(sess)
	fmt.Println("##ENTER MFA Token")
	fmt.Scanln(&mfaCode)

	params := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(900),
		SerialNumber:    aws.String(*sn),
		TokenCode:       aws.String(mfaCode),
	}
	resp, err := svc.GetSessionToken(params)

	fmt.Println(awsutil.StringValue(resp.Credentials.SessionToken))
	os.Setenv("AWS_SESSION_TOKEN", awsutil.StringValue(resp.Credentials.SessionToken))

	return sess
}

// To Get the Paramaters and Values store in SSM Parameter Store

func ssid(sess *session.Session) {
	var envvar string
	fmt.Println("Enter Required App Variables Name eg: /dev/app")
	fmt.Scanln(&envvar)
	ssmsvc := ssm.New(sess)

	param, err := ssmsvc.GetParametersByPath(&ssm.GetParametersByPathInput{
		Path:           aws.String(envvar),
		WithDecryption: aws.Bool(true),
		Recursive:      aws.Bool(true),
		MaxResults:     aws.Int64(6),
	})
	if err != nil {
		panic(err)
	}

	for _, p := range param.Parameters {
		split := strings.Split(*p.Name, "/")
		name := strings.ToUpper(split[len(split)-1])
		fmt.Println(name, ":", *p.Value)
	}

}

// To put the paramerter in parameter store.
func putpara(sess *session.Session) {
	var envname, envvalue, envtype string
	fmt.Println("Supply the Name of the parameter eg: /dev/app")
	fmt.Scanln(&envname)
	fmt.Println("Supply the Value of the parameter")
	fmt.Scanln(&envvalue)
	fmt.Println("Supply one of the listed value for Type (String,StringList,SecureString)")
	fmt.Scanln(&envtype)

	ssmsvc := ssm.New(sess)
	input, err := ssmsvc.PutParameter(&ssm.PutParameterInput{
		Name:      aws.String(envname),
		Value:     aws.String(envvalue),
		Type:      aws.String(envtype),
		Overwrite: aws.Bool(true),
	})
	if err != nil {
		panic(err)

	}
	fmt.Println("Version:", *input.Version, "added")
}

func main() {
	colorGreen := "\033[32m"
	var sess *session.Session
	getvalue()
	sess = awssess()
	var one string
	fmt.Println(string(colorGreen), "Enter 1, if you want only to see the parameters or 2 if you want to update/add new parameter")
	fmt.Scanln(&one)
	if one == "2" {
		putpara(sess)
	}
	if one == "1" {
		ssid(sess)
	}

}
