package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"transcriptions-translation-service/config"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/utils"

	"github.com/cenkalti/backoff/v4"
)

type OpenAIService struct {
	config *config.OpenAIConfig
	client *http.Client
	logger utils.Logger
}

func NewOpenAIService(cfg *config.OpenAIConfig, logger utils.Logger) *OpenAIService {
	return &OpenAIService{
		config: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.WithPrefix("[OpenAIService]"),
	}
}

const (
	defaultMaxTokens             = 2300 // for output translated text -> x1.3 of the input text, rounded to x1.5 for safety
	defaultTemperature           = 0.3
	defaultMaxCharSizePerRequest = 6300
	/*
		assuming 5 mins of talk per transcription
		-> ~770 words
		-> assuming max 2 tokens per word (1.5 on average)
		-> 1540 tokens
		-> 1540 * 4 = 6160 characters
		~1% overhead for properties that is passed along with the transcription.
	*/
)

func (s *OpenAIService) GetMaxCharSizePerRequest() int {
	return defaultMaxCharSizePerRequest
}

func (s *OpenAIService) Translate(text string, sourceLang, targetLang data.Language) (string, error) {
	jsonData := reqBuilder(s, text, sourceLang, targetLang)
	req, _ := http.NewRequest("POST", s.config.OpenAIAPIURL+"/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	respChan := make(chan *http.Response)
	errChan := make(chan error)

	go func() {
		resp, err := s.sendRequestWithRetry(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to translate text: %w", err)
			return
		}
		respChan <- resp
	}()

	select {
	case resp := <-respChan:
		defer resp.Body.Close()
		return s.parseResponse(resp)
	case err := <-errChan:
		return "", err
	}
}

func (s *OpenAIService) sendRequestWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	retryCount := 0
	operation := func() error {
		retryCount++
		resp, err = s.client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			return s.handleRequestError(retryCount, err, resp)
		}
		return nil
	}

	backoffConfig := utils.NewBackoffConfig(s.logger)
	err = backoff.RetryNotify(operation, backoffConfig.ToBackOff(), backoffConfig.Notify)
	if err != nil {
		s.logger.Error("Failed to execute request after %d retries", retryCount)
		return nil, fmt.Errorf("request failed after retries: %w", err)
	}

	return resp, nil
}

func reqBuilder(s *OpenAIService, text string, sourceLang data.Language, targetLang data.Language) []byte {
	systemPrompt := fmt.Sprintf("You are TranslateAI. Your task is to translate any user's transcriptions from %s to %s. Only translate the Arabic parts and leave any English terms, sales-related terms, or general English phrases unchanged.", string(sourceLang), string(targetLang))

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": s.config.ModelName,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": text,
			},
		},
		"max_tokens":  defaultMaxTokens,
		"temperature": defaultTemperature,
	})

	return requestBody
}

func (s *OpenAIService) parseResponse(resp *http.Response) (string, error) {
	var openAIResponse OpenAIResponse

	if err := json.NewDecoder(resp.Body).Decode(&openAIResponse); err != nil {
		s.logger.Error("error decoding response: %v", err)
		return "", fmt.Errorf("failed to decode translation response: %w", err)
	}

	if len(openAIResponse.Choices) == 0 {
		s.logger.Error("Invalid translation response format: no choices available")
		return "", errors.New("invalid translation response format: no choices available")
	}

	return openAIResponse.Choices[0].Message.Content, nil
}

func (s *OpenAIService) handleRequestError(retryCount int, err error, resp *http.Response) error {

	errorMessage := fmt.Sprintf("Attempt #%d: error executing openai request", retryCount)

	if err != nil {
		s.logger.Error("%s: %v", errorMessage, err)
		return err
	}

	s.logger.Error("%s: status code: %d", errorMessage, resp.StatusCode)
	resp.Body.Close()
	return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}
