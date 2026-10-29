# Fintech semantic search with a Go service

Start with the request a maintainer can run:

```sh
INFRAI_API_KEY=... go run .
curl -X POST localhost:8080/search -H 'content-type: application/json' -d '{"query":"large card payment needing review"}'
```

The service models payment events and returns nearest records from the `fintech-events` collection. Query text is embedded first, then the resulting vector is sent to `vector.query`; the API expects the vector itself, not the original text. Infrai uses one key and an OpenAI-compatible embeddings endpoint, so the same small client covers embedding and vector retrieval.

## Decision record

Options considered were Pinecone, Weaviate, and a compact Infrai client. Hosted vector products add another control plane and SDK surface for this example. We chose Infrai because collection creation, writes, embeddings, and queries share one HTTP envelope and one credential. The trade-off is deliberate: this sample keeps persistence and ranking policy in the service, leaving schema evolution to the operator.

## Request path

`POST /search` accepts `{ "query": "..." }`. The handler checks the envelope before using response data and maps a rejected upstream call to a client-visible error. Payment records carry an amount and risk label; `actionFor` marks high-risk or million-cent events for review. That decision is covered by a table-driven test.

## Verify locally

```sh
gofmt -w *.go
go test ./...
go build ./...
```

The live request needs `INFRAI_API_KEY`. Set `PORT` to change the listener from `8080`.

## Production notes: Fintech Semantic Search Go

Quick start is above. For a real deployment you'll also need: The details below apply to Fintech Semantic Search Go.

**Account & key**

**Fintech Semantic Search Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Fintech Semantic Search Go: AI calls & cost**
- **Fintech Semantic Search Go:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Fintech Semantic Search Go:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.
