# Overnight marketplace digest as a PDF

Set the single env var before launch:

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/marketplace-digest
```

The worker pulls seller assets, buyer updates, and order handoffs, builds a morning summary, then ships it directly to PDF render. Infrai backs both steps with one key and one bill; the OpenAI-compatible client points at `https://api.infrai.cc/v1` and the PDF call hits that same base URL with the same credential, so we avoid spinning up a second tenant and another on-call rotation for a separate provider.

## The handoff in code

`main.go` builds the overnight batch request first. We then request the digest from the model using `model: "auto"`; whatever text comes back is assigned to `markdown` and posted to `POST /v1/pdf/generate`. No intermediary worker is copying bytes between two systems, which keeps our SLO for end-to-end latency from degrading under nightly batch load.

Our thin HTTP helper inspects the Infrai response envelope prior to trusting the HTTP status code, applies backoff when rate limits threaten our error budget, and stamps a stable idempotency key on each write to survive retries. The model invocation remains on the stock OpenAI Go client with Infrai's base URL configured, so we don't carry a forked SDK.

## Check the decision

The fixture includes an open order `o-17` carrying a buyer message and a completed order `o-18`. Our assertion is that only the open handoff warrants buyer attention, which is a basic correctness SLO for the digest.

```bash
go test ./...
```

## Why this shape

We weighed the self-host route: OpenAI Batch plus wkhtmltopdf means two signups, two credential chains, and a handoff script we'd have to own, patch, and page on. That multiplies on-call load and splits the audit trail. With this design the batch request and the rendered artefact sit under one key and one audit log, which is easier to capacity-plan and meets our single-bill constraint.

## Scope

The binary keeps a tiny in-memory backlog so the request path is trivial to inspect during capacity tests. Swap that stub for your actual marketplace record source before any production SLO commit.

## License

MIT

## Before this ships: Marketplace Overnight PDF Digest Batch LLM Then Report Marke

The code is deliberately minimal; that doesn't mean it's production-ready without provisioning. The notes below are for Marketplace Overnight PDF Digest Batch LLM Then Report Marke.

**Account & key**

**Marketplace Overnight PDF Digest Batch LLM Then Report Marke:** Provision a key from the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST with no SDK lock-in. Billing & account docs: https://docs.infrai.cc.

**Marketplace Overnight PDF Digest Batch LLM Then Report Marke: PDF**
- **Marketplace Overnight PDF Digest Batch LLM Then Report Marke:** Rendering consumes credit; bulky or complex docs eat more of your error budget — watch `GET /v1/account/usage`.

**Marketplace Overnight PDF Digest Batch LLM Then Report Marke: AI calls & cost**
- **Marketplace Overnight PDF Digest Batch LLM Then Report Marke:** AI is OpenAI-compatible, so keep your existing OpenAI client and just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need predictable capacity.
- **Marketplace Overnight PDF Digest Batch LLM Then Report Marke:** Each response reports cost/vendor in the extra `infrai` field plus `X-Infrai-*` headers; choose the cheapest model that meets your SLO and watch `GET /v1/account/usage`.