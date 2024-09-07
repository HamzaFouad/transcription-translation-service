package data

type Transcription struct {
	Speaker  string `json:"speaker"`
	Time     string `json:"time"`
	Sentence string `json:"sentence"`
}

const KeyTranscriptionContext = "transcriptions_context_key"
