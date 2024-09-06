package services

type Translator interface {
	Translate(text string, sourceLang string, targetLang string) (string, error)
}
