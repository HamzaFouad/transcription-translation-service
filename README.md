# Transcriptions translation service

This is considered a microservice that translates transcriptions, currently, from `Arabic` to `English` but code be extended easily to support multiple languages with minimal configurations edits.

## Stack

go 1.23

## Running service

- you must have docker installed.
- clone the repository and navigate to the root directory
- change `.env.example` to `.env` and add your gpt-apiKey `OPENAI_API_KEY`
- run `docker build -t transcriptions-translator .` to build the image locally
- run the image `docker run -p 9000:9000 --env-file .env --name transcriptions-translator_container transcriptions-translator`

## Endpoints

```curl
POST {baseUrl}/translate
```

### Example

```bash
curl --location 'http://localhost:9000/translate' \
--header 'Content-Type: application/json' \
--data '[
    {
        "speaker": "John",
        "time": "00:00:04",
        "sentence": "Hello Maria."
    },
    {
        "speaker": "Maria",
        "time": "00:00:09",
        "sentence": "صباح الخير"
    }
]'
```

and the expected response will be as following

```json
[
    {
        "speaker": "John",
        "time": "00:00:04",
        "sentence": "Hello Maria."
    },
    {
        "speaker": "Maria",
        "time": "00:00:09",
        "sentence": "Good morning"
    }
]
```

---

## Checklist

- [x] My first ever golang project :D
- [x] Implement service that handles basic scenarios, like sending multi-transcriptions array with moderate sizes.
- [x] Batch small transcriptions into moderate batches and translate them accordingly
- [x] Handle batch translations concurrently with goroutines.
- [x] Add retry policy and handle timeouts on openai-api for failed translations.
- [x] Log errors.
- [ ] Handle failed translations... right now, I return error to the user if any batch failed.. and I don't also fail fast, so I wait for the rest of the translations then return the error
- [ ] Fail fast. (as mentioned above.)
- [ ] split large transcriptions into smaller ones..
    -> had with this, trying to split transcriptions with punctuations but lots of times there none.. so it didn't give any good results, imo, we'll eventually need to use a NLP model that make the segmentation so the split strings still have solid meaning and don't lose context to help LLM translate more efficiently..etc.
- [ ] testing-pipeline
- [ ] rate limiting, from both sides, our service, and openai requests side, need to check their documentations for their api rate limits..

## Flow

- `/translate`
- validate input
- extract sentences that needs to be translated into a separate list, group them into moderate batches.
- every batch opens a goroutine and calls openai to translate the batch, having its own retry policy if there is an  translate these batches concurrently, while preserving the batches orders. [every batch opens a separate thread with its own retries]
- await all goroutines and assemble all these transcriptions back into results array with the same order as user's request.

## Decisions

**Chunks sizes**

- I assumed on average that `every transcription block is for around ~5 mins of speech`, imo this is plenty as usually transcriptions are split into more transcription blocks based on stops...etc.
- On average, I assumed human speaks around 150 words per minute -> ~770 words per minute
- Assuming 2 tokens are generated per word (it is 1.5 on average) -> ~ 1540 tokens
- tokens ~= numberOfCharacters / 4 as per openAI docs -> so ~6160 characters
and assuming ~1% overhead for special chars that is sent along with the text.
- So, transcriptions blocks are assumed to be around `6300 characters`, so we group transcriptions blocks that don't exceed this limitation to send translation batches to openai.

**Gpt-model**

- `gpt-4o-mini` provides highly accurate translations for block sizes around the size we use and is considered very cost-efficient.

---

you can checkout some examples of requests I tested here.

https://docs.google.com/spreadsheets/d/1unKsanRhe48-Bj44TxYxLdx0Ck-v4_VmPTDRWg_nBXg/edit?usp=sharing 
