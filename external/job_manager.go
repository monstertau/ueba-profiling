package external

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"ueba-profiling/config"
	"ueba-profiling/view"
)

type JobHub struct {
	endpoint string
	timeout  time.Duration
}

func NewJobHub() *JobHub {
	return &JobHub{
		endpoint: config.GlobalConfig.EndpointGetJobs,
		timeout:  1 * time.Minute,
	}
}

func (j *JobHub) GetJobs() ([]*view.JobConfig, error) {
	client := http.Client{
		Timeout: time.Minute,
	}
	var jobs []*view.JobConfig
	req, err := http.NewRequest(http.MethodGet, j.endpoint, nil)

	if err != nil {
		return nil, fmt.Errorf("in NewRequest at endpoint %s: %s", j.endpoint, err)

	}
	req.Header.Add("Accept", "application/json")

	// make requests
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot run the request at endpoint %s: %s", j.endpoint, err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read the response at endpoint %s: %s", j.endpoint, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, response: '%s'\n", resp.StatusCode, string(respBody))
	}
	if err := json.Unmarshal(respBody, &jobs); err != nil {
		return nil, fmt.Errorf("unexpected response data at endpoint %s: %s", j.endpoint, err)
	}
	return jobs, nil
}
