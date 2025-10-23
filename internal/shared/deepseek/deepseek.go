package deepseek

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

const deepseekBaseURL = "https://api.deepseek.com/chat/completions"

type Client struct {
	apiKey string
}

func NewClient(apikey string) *Client {
	return &Client{
		apiKey: apikey,
	}
}

func (c *Client) Chat(ctx context.Context, request ChatRequest) (*ChatResponse, error) {

	// get json
	data, err := json.Marshal(request)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal JSON")
		return nil, err
	}

	// create request with headers api berarer
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, deepseekBaseURL, bytes.NewReader(data))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP request")
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	log.Info().Msgf("header %s", httpReq.Header)

	// get the response
	response, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send HTTP request")
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		log.Error().Msgf("Failed to send HTTP request, status code %d", response.StatusCode)
	}
	defer response.Body.Close()

	// TODO should it be unmarshaled into a struct?
	var chatResponse ChatResponse
	respBody, err := io.ReadAll(response.Body)
	log.Info().Msgf("resp body %s", respBody)
	if err := json.Unmarshal(respBody, &chatResponse); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal JSON")
		return nil, err
	}

	return &chatResponse, nil
}
