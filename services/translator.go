package services

import "transcriptions-translation-service/data"

type Translator interface {
	Translate(text string, sourceLang, targetLang data.Language) (string, error)
}
