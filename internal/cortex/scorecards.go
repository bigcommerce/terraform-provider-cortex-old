package cortex

import (
	"context"
	"errors"
	"github.com/dghubble/sling"
	"gopkg.in/yaml.v3"
	"strings"
)

type ScorecardsClientInterface interface {
	Get(ctx context.Context, tag string) (*Scorecard, error)
	List(ctx context.Context, params *ScorecardListParams) (*ScorecardsResponse, error)
	Upsert(ctx context.Context, req UpsertScorecardRequest) (*Scorecard, error)
	Delete(ctx context.Context, tag string) error
}

type ScorecardsClient struct {
	client *HttpClient
}

var _ ScorecardsClientInterface = &ScorecardsClient{}

func (c *ScorecardsClient) Client() *sling.Sling {
	return c.client.client
}

/***********************************************************************************************************************
 * Types
 **********************************************************************************************************************/

// Scorecard is the nested response object that is typically returned from the scorecards endpoints.
type Scorecard struct {
	Tag         string          `json:"tag" yaml:"tag"`
	Name        string          `json:"name,omitempty" yaml:"name,omitempty"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty"`
	IsDraft     bool            `json:"is_draft" yaml:"is_draft"`
	Rules       []ScorecardRule `json:"rules" yaml:"rules"`
}

type ScorecardRule struct {
	Title          string `json:"title" yaml:"title"`
	Description    string `json:"description" yaml:"description"`
	Expression     string `json:"expression" yaml:"expression"`
	FailureMessage string `json:"failure_message" yaml:"failureMessage"`
	LevelName      string `json:"level_name" yaml:"levelName"`
	Weight         int    `json:"weight" yaml:"weight"`
}

type ScorecardLadder struct {
	Levels []ScorecardLevel `json:"levels" yaml:"levels"`
}

type ScorecardLevel struct {
	Name        string `json:"name" yaml:"name"`
	Rank        int    `json:"rank" yaml:"rank"`
	Color       string `json:"color" yaml:"color"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

/***********************************************************************************************************************
 * GET /api/v1/scorecards/:tag
 **********************************************************************************************************************/

// GetScorecardResponse is the generic root response object for scorecards.
type GetScorecardResponse struct {
	Scorecard *Scorecard `json:"scorecard"`
}

func (c *ScorecardsClient) Get(ctx context.Context, tag string) (*Scorecard, error) {
	scorecardResponse := &GetScorecardResponse{
		Scorecard: &Scorecard{},
	}
	apiError := &ApiError{}
	response, err := c.Client().Get(BaseUris["scorecards"]+tag).Receive(scorecardResponse, apiError)
	if err != nil {
		return scorecardResponse.Scorecard, errors.New("could not get scorecard: " + err.Error())
	}

	err = c.client.handleResponseStatus(response, apiError)
	if err != nil {
		return scorecardResponse.Scorecard, errors.Join(errors.New("Failed getting scorecard: "), err)
	}

	return scorecardResponse.Scorecard, nil
}

/***********************************************************************************************************************
 * GET /api/v1/scorecards
 **********************************************************************************************************************/

// ScorecardListParams are the query parameters for the GET /v1/scorecards endpoint.
type ScorecardListParams struct {
}

// ScorecardsResponse is the response from the GET /v1/scorecards endpoint.
type ScorecardsResponse struct {
	Scorecards []Scorecard `json:"scorecards"`
}

// List retrieves a list of scorecards based on a query.
func (c *ScorecardsClient) List(ctx context.Context, params *ScorecardListParams) (*ScorecardsResponse, error) {
	scorecardsResponse := &ScorecardsResponse{}
	apiError := &ApiError{}

	response, err := c.Client().Get(BaseUris["scorecards"]).QueryStruct(params).Receive(scorecardsResponse, apiError)
	if err != nil {
		return nil, errors.New("could not get scorecards: " + err.Error())
	}

	err = c.client.handleResponseStatus(response, apiError)
	if err != nil {
		return nil, err
	}

	return scorecardsResponse, nil
}

/***********************************************************************************************************************
 * POST /api/v1/scorecards/descriptor
 **********************************************************************************************************************/

// UpsertScorecardRequest is the request object that is a struct representation of the Scorecard descriptor YAML file.
// We do this to allow for easier creation of scorecards in Go or Terraform, rather than having to pass through a YAML
// file or string, and figure out how to tell Terraform to "compare" that data.
type UpsertScorecardRequest struct {
	Tag         string          `yaml:"tag"`
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	IsDraft     bool            `yaml:"is_draft"`
	Rules       []ScorecardRule `yaml:"rules"`
	Ladder      ScorecardLadder `yaml:"ladder"`
}

type UpsertScorecardResponse struct {
	Scorecard *Scorecard `json:"scorecard"`
}

func (c *ScorecardsClient) Upsert(ctx context.Context, req UpsertScorecardRequest) (*Scorecard, error) {
	upsertScorecardResponse := &UpsertScorecardResponse{
		Scorecard: &Scorecard{},
	}
	apiError := &ApiError{}

	// The API requires submitting the request as YAML, so we need to marshal it first.
	bytes, err := yaml.Marshal(req)
	if err != nil {
		return upsertScorecardResponse.Scorecard, errors.New("could not marshal yaml: " + err.Error())
	}
	body := strings.NewReader(string(bytes))
	response, err := c.Client().Post(BaseUris["scorecards"]).Body(body).Receive(upsertScorecardResponse, apiError)
	if err != nil {
		return upsertScorecardResponse.Scorecard, errors.New("could not upsert scorecard: " + err.Error())
	}

	err = c.client.handleResponseStatus(response, apiError)
	if err != nil {
		return upsertScorecardResponse.Scorecard, err
	}

	return upsertScorecardResponse.Scorecard, nil
}

/***********************************************************************************************************************
 * DELETE /api/v1/scorecards/:tag - Delete a scorecard
 **********************************************************************************************************************/

type DeleteScorecardResponse struct{}

func (c *ScorecardsClient) Delete(ctx context.Context, tag string) error {
	scorecardResponse := &DeleteScorecardResponse{}
	apiError := &ApiError{}

	response, err := c.Client().Delete(BaseUris["scorecards"]+tag).Receive(scorecardResponse, apiError)
	if err != nil {
		return errors.New("could not delete scorecard: " + err.Error())
	}

	err = c.client.handleResponseStatus(response, apiError)
	if err != nil {
		return err
	}

	return nil
}
