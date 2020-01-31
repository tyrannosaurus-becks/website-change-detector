package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
)

const (
	EnvVarTwilioAccountSID = "TWILIO_ACCOUNT_SID"
	EnvVarTwilioAuthToken  = "TWILIO_AUTH_TOKEN"

	twilioAPIBaseURL = "https://api.twilio.com/2010-04-01"
)

var (
	// Command-line flags.
	pageToCheck = flag.String("url", "", `ex: "https://google.com"`)
	matchPhrase = flag.String("phrase", "", `ex: "Hello World"`)
	to          = flag.String("to", "", `recipient's phone number, ex: "503-123-4567"`)
	from        = flag.String("from", "", `sender's phone number, ex: "503-123-4567", should exist in Twilio`)
	frequency   = flag.Int("frequency", 60, `ex: "60", how frequently in seconds the web page should be checked, defaults to 60`)
	dryRun      = flag.Bool("dryrun", false, `if included, indicates that notifications should be logged instead of sent'`)

	// Logger for tool use.
	logger = hclog.Default()
)

func main() {
	flag.Parse()

	twilioAccountSID := os.Getenv(EnvVarTwilioAccountSID)
	if twilioAccountSID == "" {
		if !*dryRun {
			logger.Error(EnvVarTwilioAccountSID + " not set in environment")
			os.Exit(1)
		}
	}
	twilioAuthToken := os.Getenv(EnvVarTwilioAuthToken)
	if twilioAuthToken == "" {
		if !*dryRun {
			logger.Error(EnvVarTwilioAuthToken + " not set in environment")
			os.Exit(1)
		}
	}
	if pageToCheck == nil || *pageToCheck == "" {
		logger.Error(`"url" flag must be included`)
		os.Exit(1)
	}
	if matchPhrase == nil || *matchPhrase == "" {
		logger.Error(`"phrase" flag must be included`)
		os.Exit(1)
	}
	if to == nil || *to == "" {
		logger.Error(`"to" flag must be included`)
		os.Exit(1)
	}
	if from == nil || *from == "" {
		logger.Error(`"from" flag must be included`)
		os.Exit(1)
	}
	checkFrequency := time.Second * time.Duration(*frequency)

	for i := 0; true; i++ {
		if i != 0 {
			<-time.NewTimer(checkFrequency).C
		}
		logger.Info("checking...")
		hasPhrase, err := checkPage()
		if err != nil {
			logger.Warn(fmt.Sprintf("could not get page: %s", err))
			continue
		}
		if hasPhrase {
			continue
		}
		logger.Info(fmt.Sprintf("%q no longer has %q", *pageToCheck, *matchPhrase))
		if !*dryRun {
			if err := notify(twilioAccountSID, twilioAuthToken); err != nil {
				logger.Warn(fmt.Sprintf("could not send notification: %s", err))
				continue
			}
		}
		return
	}
}

func checkPage() (bool, error) {
	resp, err := http.Get(*pageToCheck)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Warn(fmt.Sprintf("unable to close page response body due to %s", err))
		}
	}()
	if resp.StatusCode != 200 {
		return false, errors.New(resp.Status)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(b), *matchPhrase), nil
}

func notify(twilioAccountSID, twilioAuthToken string) error {
	endpoint := "/Accounts/" + twilioAccountSID + "/Messages.json"
	method := http.MethodPost

	v := url.Values{}
	v.Set("To", *to)
	v.Set("From", *from)
	v.Set("Body", fmt.Sprintf("%q no longer has %q", *pageToCheck, *matchPhrase))
	rb := *strings.NewReader(v.Encode())

	req, err := http.NewRequest(method, twilioAPIBaseURL+endpoint, &rb)
	if err != nil {
		return err
	}
	req.SetBasicAuth(twilioAccountSID, twilioAuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Warn(fmt.Sprintf("unable to close SMS response body due to %s", err))
		}
	}()
	if resp.StatusCode != 201 {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%d: %s", resp.StatusCode, b)
	}
	return nil
}
