package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"transcriptions-translation-service/config"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/utils"
)

type OpenAIService struct {
	config *config.OpenAIConfig
	client *http.Client
	logger utils.Logger
}

func NewOpenAIService(cfg *config.OpenAIConfig, logger utils.Logger) *OpenAIService {
	return &OpenAIService{
		config: cfg,
		client: &http.Client{},
		logger: logger.WithPrefix("[OpenAIService]"),
	}
}

const (
	defaultMaxTokens             = 2300 // for output translated text -> x1.3 of the input text, rounded to x1.5 for safety
	defaultTemperature           = 0.3
	DefaultMaxCharSizePerRequest = 6300
	/*
		assuming 5 mins of talk per transcription
		-> ~770 words
		-> assuming max 2 tokens per word (1.5 on average)
		-> 1540 tokens
		-> 1540 * 4 = 6160 characters
		~1% overhead for properties that is passed along with the transcription.
	*/
)

func (s *OpenAIService) Translate(text string, sourceLang, targetLang data.Language) (string, error) {

	jsonData := reqBuilder(s, text, sourceLang, targetLang)
	req, _ := http.NewRequest("POST", s.config.OpenAIAPIURL+"/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("error executing request: %v", err)
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("failed to translate text, status code: %d", resp.StatusCode)
		return "", fmt.Errorf("failed to translate text, status code: %d", resp.StatusCode)
	}

	return s.parseResponse(resp)
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
