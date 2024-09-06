package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"transcriptions-translation-service/config"
	"transcriptions-translation-service/data"
)

type OpenAIService struct {
	config *config.OpenAIConfig
	client *http.Client
}

func NewOpenAIService(cfg *config.OpenAIConfig) *OpenAIService {
	return &OpenAIService{
		config: cfg,
		client: &http.Client{},
	}
}

const (
	defaultMaxTokens   = 2500
	defaultTemperature = 0.3
)

func (s *OpenAIService) Translate(text string, sourceLang, targetLang data.Language) (string, error) {

	jsonData := reqBuilder(s, text, sourceLang, targetLang)
	req, _ := http.NewRequest("POST", s.config.OpenAIAPIURL+"/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
		return "", fmt.Errorf("failed to decode translation response: %w", err)
	}

	if len(openAIResponse.Choices) == 0 {
		return "", errors.New("invalid translation response format: no choices available")
	}

	return openAIResponse.Choices[0].Message.Content, nil
}
