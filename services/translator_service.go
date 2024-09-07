package services

import (
	"fmt"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/utils"
)

type Result struct {
	Index           int
	TranslatedBatch []string
	Err             error
}

func ProcessBatchesConcurrently(batches [][]string, translator Translator, sourceLang, targetLang data.Language, logger utils.Logger) ([]string, error) {
	resultsChan := make(chan Result)
	defer close(resultsChan)

	for index, batch := range batches {
		go TranslateBatch(index, batch, translator, sourceLang, targetLang, resultsChan)
	}

	return CollectResults(len(batches), resultsChan, logger)
}

func TranslateBatch(index int, batch []string, translator Translator, sourceLang, targetLang data.Language, resultsChan chan<- Result) {
	serializedBatch, err := utils.SerializeToString(batch)
	if err != nil {
		resultsChan <- Result{Index: index, Err: fmt.Errorf("failed to serialize batch: %v", err)}
		return
	}

	translatedText, err := translator.TranslateAsync(serializedBatch, sourceLang, targetLang)
	if err != nil {
		resultsChan <- Result{Index: index, Err: err}
		return
	}

	if !utils.IsValidJSON(translatedText) {
		resultsChan <- Result{Index: index, Err: fmt.Errorf("invalid JSON format received: %s", translatedText)}
		// potentially here we can store the failed batch in a queue and handle it accordingly
		return
	}

	translatedBatch, err := utils.DeserializeTranslatedText(translatedText)
	if err != nil {
		resultsChan <- Result{Index: index, Err: err}
		return
	}

	resultsChan <- Result{Index: index, TranslatedBatch: translatedBatch}
}

func CollectResults(batchCount int, resultsChan <-chan Result, logger utils.Logger) ([]string, error) {
	results := make([][]string, batchCount)
	translatedTexts := make([]string, 0)

	for i := 0; i < batchCount; i++ {
		res := <-resultsChan
		if res.Err != nil {
			logger.Error("Translation error for batch: %v", res.Err)
			return nil, res.Err
		}
		results[res.Index] = res.TranslatedBatch
	}

	// Flatten the results into a single slice
	for _, translatedBatch := range results {
		translatedTexts = append(translatedTexts, translatedBatch...)
	}

	return translatedTexts, nil
}
