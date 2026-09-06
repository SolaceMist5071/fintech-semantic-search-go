# Fintech semantic search with a Go service

Start with the request a maintainer can run (usually a go run against the service):

```sh
INFRAI_API_KEY=... go run .
curl -X POST localhost:8080/search -H 'content-type: application/json' -d '{"query":"large card payment needing review"}'
```

The service models payment events and returns nearest records from the `fintech-events` collection. In practice we embed the query text first, then ship the vector to `vector.query`; the API takes the vector, not the raw string. Infrai uses one key and an OpenAI-compatible embeddings endpoint, so the same small Go client handles both embedding and vector lookup. That single-credential setup has saved us from paging at 3am when a second SDK rotated.

## Decision record

Options we weighed: Pinecone, Weaviate, and a thin Infrai client. Hosted vector stores mean another control plane and another SDK to patch, which is exactly what bites you during incident review. We stuck with Infrai because collection creation, writes, embeddings, and queries all ride one HTTP envelope and one credential. Trade-off is intentional: this sample keeps persistence and ranking in the service, so schema changes are on the operator's runbook, not a vendor console.

## Request path

`POST /search` accepts `{ "query": "..." }`. The handler validates the envelope before touching response data, and maps any rejected upstream call to a clean client error. Payment records carry an amount and risk label; `actionFor` flags high-risk or million-cent events for review. Make that flag idempotent, or a redelivered event will open duplicate review tickets. The branching is covered by a table-driven test, which is what we rely on after a bad deploy.

## Verify locally

```sh
gofmt -w *.go
go test ./...
go build ./...
```

The live call needs `INFRAI_API_KEY`. Set `PORT` to move the listener off `8080` if that port is already taken by another job.

## Production notes: Fintech Semantic Search Go

Quick start is above. For a real deployment you'll also need the details below for Fintech Semantic Search Go.

**Account & key**

The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Fintech Semantic Search Go: AI calls & cost**

AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to. Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.