package fresh_container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	Server string
}

func NewClient(server string) Client {
	return Client{
		Server: server,
	}
}

func (c *Client) EvalUpgrade(image, constraint, tagPrefix string) (RemoteEvaluationResponse, error) {
	u, err := url.Parse(c.Server)
	if err != nil {
		return RemoteEvaluationResponse{}, err
	}

	u.Path = "/api/v1/check"
	q := u.Query()
	q.Add("image", image)
	q.Add("constraint", constraint)
	if tagPrefix != "" {
		q.Add("tagPrefix", tagPrefix)
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return RemoteEvaluationResponse{}, err
	}

	log.WithFields(log.Fields{
		"image":      image,
		"constraint": constraint,
		"tagPrefix":  tagPrefix,
		"resp-code":  resp.Status,
		"headers":    resp.Header,
	}).Debug("Remote evaluation response")

	switch resp.StatusCode {
	case http.StatusAccepted:
		return RemoteEvaluationResponse{
			Pending:      true,
			server:       c.Server,
			jobStatusUrl: resp.Header.Get("Location"),
		}, nil
	case http.StatusOK:
		var evaluation ImageUpgradeEvaluationResponse

		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&evaluation)
		if err != nil {
			return RemoteEvaluationResponse{}, err
		}

		return RemoteEvaluationResponse{
			server:   c.Server,
			Pending:  false,
			Response: evaluation,
		}, nil
	default:
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return RemoteEvaluationResponse{}, err
		}
		err = fmt.Errorf(
			"%s - Response code: %s - Body %s",
			u.String(),
			resp.Status,
			body)
		return RemoteEvaluationResponse{}, err
	}
}

type RemoteEvaluationResponse struct {
	server       string
	Pending      bool
	Response     ImageUpgradeEvaluationResponse
	jobStatusUrl string
}

func (re *RemoteEvaluationResponse) IsReady() (bool, error) {
	if !re.Pending {
		return true, nil
	}

	log.WithFields(log.Fields{
		"server":       re.server,
		"jobStatusUrl": re.jobStatusUrl,
	}).Debug("Poll job status")

	u, err := url.Parse(re.server)
	if err != nil {
		return false, err
	}
	u.Path = re.jobStatusUrl

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(u.String())
	if err != nil {
		return false, err
	}

	switch resp.StatusCode {
	case http.StatusSeeOther:
		err := re.fetchEvaluation(resp.Header.Get("Location"))
		if err != nil {
			return false, err
		}
		return true, nil
	case http.StatusOK:
		return false, nil
	default:
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		err = fmt.Errorf(
			"%s - Response code: %s - Body %s",
			u.String(),
			resp.Status,
			body)
		return false, err
	}
}

func (re *RemoteEvaluationResponse) fetchEvaluation(path string) error {
	u, err := url.Parse(re.server)
	if err != nil {
		return err
	}
	u.Path = path

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var evaluation ImageUpgradeEvaluationResponse

		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&evaluation)
		if err != nil {
			return err
		}
		re.Response = evaluation
		return nil
	default:
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = fmt.Errorf(
			"%s - Response code: %s - Body %s",
			u.String(),
			resp.Status,
			body)
		return err
	}
}
