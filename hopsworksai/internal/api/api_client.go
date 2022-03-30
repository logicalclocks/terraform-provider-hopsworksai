package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const host = "https://api.hopsworks.ai"

type ResponseWithValidator interface {
	validate() error
}

type APIHandler interface {
	doRequest(ctx context.Context, method string, endpoint string, body io.Reader, response ResponseWithValidator) error
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HopsworksAIClient struct {
	Client     HttpClient
	UserAgent  string
	ApiKey     string
	ApiVersion string
}

func (a *HopsworksAIClient) doRequest(ctx context.Context, method string, endpoint string, body io.Reader, response ResponseWithValidator) error {
	url := host + endpoint
	tflog.Debug(ctx, method+" "+url)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", a.UserAgent)
	req.Header.Set("x-api-key", a.ApiKey)
	req.Header.Set("hopsai-api-version", a.ApiVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		bodyBytes, respErr := ioutil.ReadAll(resp.Body)
		if respErr != nil {
			return fmt.Errorf("the API token provided does not have access to hopsworks.ai, verify the token you specified matches the token hopsworks.ai created")
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("the API token provided does not have access to hopsworks.ai, verify the token you specified matches the token hopsworks.ai created:\n\t%s", bodyString)
	}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("failed to decode json, resp: %s, path: %s err: %s", resp.Status, url, err)
	}

	tflog.Debug(ctx, fmt.Sprintf("response struct: %#v", response))

	if err := response.validate(); err != nil {
		return err
	}
	return nil
}
