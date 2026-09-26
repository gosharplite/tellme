# Model API Specs — How To Get Them Live

> **Status: descriptive reference, not truth.** On any conflict, [`specs/truth/**`](../specs/truth)
> wins (it describes *tellme's* observed wire behaviour). This file only records **where to fetch the
> upstream vendor specs** and **how to reproduce a call**, so a DeepSeek/Gemini issue can be checked
> against the current contract instead of a stale memory.
>
> **Last verified live: 2026-09-26** (all URLs below were fetched; status/content-type/bytes recorded).

## 0. The one rule

When a DeepSeek or Gemini issue appears, do these **in order** — do not conclude from memory:

1. **Test egress** (is the sandbox online at all?). `curl -sS -o /dev/null -w '%{http_code}\n' <url>`
2. **Fetch the live spec** (§2 / §3 below) and compare the failing field/shape to the adapter.
3. **Reproduce the call** with a real credential (§5) — reading the spec is necessary, not sufficient;
   the defect is often in *tellme's* hand-rolled adapter, config expansion, or auth, not the API.

Access facts established on this host: general HTTPS egress works (no proxy needed); DeepSeek and
Google API hosts are reachable; credentials for both exist (§5). The repo's *tests* are deliberately
offline/hermetic — that does **not** forbid an operator/agent from checking a spec or pinging an API.

---

## 1. Quick reference

| Vendor | Live docs | Machine-readable spec | Live call | Wire tellme speaks |
| --- | --- | --- | --- | --- |
| **DeepSeek** | [api-docs.deepseek.com](https://api-docs.deepseek.com/) (HTML, server-rendered) | ❌ none (no OpenAPI/Swagger) | ✅ `DEEPSEEK_API_KEY` | OpenAI Chat Completions + DeepSeek deltas |
| **Gemini — public Gemini API** | [ai.google.dev/gemini-api/docs](https://ai.google.dev/gemini-api/docs) | ✅ **discovery doc** (§3.1) | ✅ (API key) | `generativelanguage.googleapis.com` |
| **Gemini — Vertex AI (what tellme uses)** | [cloud.google.com/vertex-ai/…](https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/inference) | ✅ **discovery doc** (§3.2) | ✅ service-account | `aiplatform.googleapis.com/…:generateContent` |

> **DeepSeek is the exception:** it publishes **no machine-readable spec** — only HTML docs. Its API is
> self-described as *"compatible with OpenAI/Anthropic"*, so the authoritative wire spec is **OpenAI's
> Chat Completions** spec **plus DeepSeek's deltas** (thinking toggle, `user_id`, `reasoning_content`).

---

## 2. DeepSeek

### 2.1 Docs (HTML; there is **no** OpenAPI)

| Page | URL | Verified |
| --- | --- | --- |
| Landing / first call | `https://api-docs.deepseek.com/` | 200, `text/html` |
| API reference | `https://api-docs.deepseek.com/api/deepseek-api` | 200, `text/html` |
| Create Chat Completion | `https://api-docs.deepseek.com/api/create-chat-completion` | 200, `text/html` (~143 KB) |
| Reasoning model guide | `https://api-docs.deepseek.com/guides/reasoning_model` | 200, `text/html` |
| Pricing / multi-round | `/quick_start/pricing`, `/guides/multi_round_chat` | 302 → 200 |

The pages are **Docusaurus server-rendered** — the spec text *is* in the raw HTML (grep-verified:
`reasoning_content` ×9, `user_id` ×7, `frequency_penalty`, `chat/completions` all present), so
`curl … | grep`/`pandoc` works without a browser.

```bash
# fetch a DeepSeek page and pull out concrete field mentions
curl -sL https://api-docs.deepseek.com/api/create-chat-completion -o /tmp/ds.html
grep -o -iE 'reasoning_content|user_id|thinking|frequency_penalty|logprobs' /tmp/ds.html | sort | uniq -c
```

**Dead ends (verified):** `…/openapi.json` and `…/swagger.json` return the SPA **HTML fallback**
(`text/html`, ~48 KB) — **not** JSON. Do not treat them as a spec. `api.deepseek.com/openapi.json`
→ 401.

### 2.2 The OpenAI-compatible substrate

Because DeepSeek mirrors OpenAI's Chat Completions, the *authoritative* field/shape spec lives here:

```bash
curl -sL https://raw.githubusercontent.com/openai/openai-openapi/master/openapi.yaml \
  -o /tmp/openai-openapi.yaml   # 200, text/plain, ~2.0 MB
```

Read OpenAI for the *shape*; read the DeepSeek pages for the *deltas*.

### 2.3 tellme's DeepSeek wire

- Base URL: `https://api.deepseek.com` (tellme appends `/chat/completions`).
- Adapter: [`internal/infrastructure/llm/openai/client.go`](../internal/infrastructure/llm/openai/client.go)
  (the OpenAI-compatible family — labels `openai` / `deepseek` / `kimi`).
- Deltas tellme relies on: `finish_reason:"length"` output-cap guard (round 030),
  `usage.prompt_tokens_details.cached_tokens` / `completion_tokens_details.reasoning_tokens`
  (round 072), and (config-declared) the thinking toggle / `user_id`.
- The label→family map is single-owned in [`internal/infrastructure/llm/factory.go`](../internal/infrastructure/llm/factory.go).

---

## 3. Gemini

Tellme drives the **Vertex AI** surface (`aiplatform.googleapis.com/.../publishers/google/models/...:generateContent`),
but the **public** Gemini API is the sibling and is the cleaner spec to read. Both publish a
**discovery document** — the real, machine-readable JSON schema (not rendered prose).

### 3.1 Public Gemini API discovery doc

```bash
# v1beta (~382 KB) and v1 (~236 KB) — both application/json
curl -sL 'https://generativelanguage.googleapis.com/$discovery/rest?version=v1beta' -o /tmp/gemini.v1beta.json
curl -sL 'https://generativelanguage.googleapis.com/$discovery/rest?version=v1'     -o /tmp/gemini.v1.json
```

```bash
# extract the request/response schemas (verified keys)
python3 - <<'PY'
import json; d=json.load(open('/tmp/gemini.v1beta.json'))
print('revision:', d['revision'])                       # e.g. 20260925
print('request :', list(d['schemas']['GenerateContentRequest']['properties']))
print('response:', list(d['schemas']['GenerateContentResponse']['properties']))
print('candidate:', list(d['schemas']['Candidate']['properties']))
PY
```

Verified: `GenerateContentRequest` → `systemInstruction, contents, toolConfig, generationConfig,
cachedContent, …`; `GenerateContentResponse` → `promptFeedback, candidates, usageMetadata,
modelVersion, responseId, …`; `models` methods include `generateContent`, `streamGenerateContent`,
`countTokens`, `embedContent`, …; a `UsageMetadata` schema is present.

### 3.2 Vertex AI discovery doc (the surface tellme uses)

```bash
curl -sL 'https://aiplatform.googleapis.com/$discovery/rest?version=v1' -o /tmp/vertex.v1.json   # 200, ~3.8 MB
# NOTE: version=v1beta on aiplatform returns 404 (verified) — use v1.
```

Discover other Google APIs via the index:

```bash
curl -sL 'https://www.googleapis.com/discovery/v1/apis?preferred=true' | \
  python3 -c "import sys,json;print([(i['name'],i['version']) for i in json.load(sys.stdin)['items'] if i['name'] in ('aiplatform','generativelanguage')])"
```

### 3.3 Docs pages (human-readable)

- `https://ai.google.dev/api/generate-content` (200) — public Gemini `generateContent` reference.
- `https://ai.google.dev/gemini-api/docs` (200).
- `https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/inference` (301 → 200) — Vertex inference reference.

### 3.4 tellme's Gemini wire

- Endpoint: `<configured URL>/<model>:generateContent` — e.g.
  `https://aiplatform.googleapis.com/v1/projects/<PROJECT>/locations/global/publishers/google/models/<model>:generateContent`.
- Adapter: [`internal/infrastructure/llm/gemini/client.go`](../internal/infrastructure/llm/gemini/client.go)
  (stdlib `net/http` only — **no Google SDK**).
- Auth: [`internal/infrastructure/llm/gemini/auth.go`](../internal/infrastructure/llm/gemini/auth.go) —
  service-account JSON → JWT-RS256 assertion → token at `token_uri`
  (default `https://oauth2.googleapis.com/token`), scope `…/auth/cloud-platform`, cached in memory.
- Fields tellme relies on: `contents`/`parts`, `role:"model"` `functionCall`, `role:"user"`
  `functionResponse`, `inlineData`/`mimeType`, `systemInstruction`, `generationConfig`, and
  `usageMetadata.cachedContentTokenCount` / `thoughtsTokenCount` (round 072), plus the replayed
  `thoughtSignature`.

---

## 4. Reconciling an adapter against the live spec

```bash
# 1. refresh the spec
curl -sL 'https://generativelanguage.googleapis.com/$discovery/rest?version=v1beta' -o /tmp/gemini.json
# 2. list what tellme emits / reads
grep -nE 'json:"|payload\[|map\[string\]any' internal/infrastructure/llm/gemini/client.go
# 3. compare each key against the discovery schema (e.g.)
python3 - <<'PY'
import json; d=json.load(open('/tmp/gemini.json'))
s=d['schemas']
for k in ['GenerateContentRequest','Content','Part','Blob','FunctionCall','FunctionResponse','UsageMetadata','GenerationConfig']:
    print(k, '->', list(s.get(k,{}).get('properties',{}))[:14])
PY
```

For DeepSeek, the same shape check uses `/tmp/openai-openapi.yaml` (§2.2) + the DeepSeek HTML deltas.

---

## 5. Credentials & reproducing a call

**Never print a secret.** Check *presence* only:

```bash
for v in DEEPSEEK_API_KEY GOOGLE_APPLICATION_CREDENTIALS GOOGLE_PROJECT_ID; do
  printenv "$v" >/dev/null && echo "$v = SET" || echo "$v = unset"
done
```

| Vendor | Credential source (resolved by `$TELL_ME_HOME/configs/*.yaml` `${VAR}` expansion) |
| --- | --- |
| DeepSeek | `${DEEPSEEK_API_KEY}` (inline bearer) |
| Gemini/Vertex | a **service-account JSON** path — e.g. `${NIFFLER_HOME}/secrets/key.json`; tellme mints the OAuth2 token itself |

```bash
# DeepSeek: minimal live ping (env var reference — the value is never echoed)
curl -sS https://api.deepseek.com/chat/completions \
  -H "Authorization: Bearer $DEEPSEEK_API_KEY" -H 'Content-Type: application/json' \
  -d '{"model":"deepseek-flash","messages":[{"role":"user","content":"ping"}],"max_tokens":8}' | head -c 400

# Gemini/Vertex: prefer the tellme binary (it mints the token from the SA key, no token on the CLI):
tellme -c "$TELL_ME_HOME/configs/butler.yaml" -r "ping"     # spends tokens — operator-gated
```

> **Live calls cost money and are operator-gated.** Prefer the **read-only** spec fetches (§2/§3) and
> the **hermetic** E2E fakes for regression; use a live ping only to confirm a suspected transport
> change. Tellme encodes only the single non-streaming **turn-loop subset** — a spec says what is
> *allowed*, not what tellme *uses*.

---

## 6. Checklist when an issue arises

- [ ] Egress OK? (`curl -o /dev/null -w '%{http_code}'`)
- [ ] Spec source refreshed (§2 DeepSeek HTML / §3 Gemini discovery)?
- [ ] Field/shape diffed against the adapter (`grep json:"…"`)?
- [ ] If live: credential present (presence-only) and the call reproduced?
- [ ] Fix landed with an **executable witness** (per `axb-tasks` Claim→Witness; the ADR 0006 obligation)
      so the regression is caught without a live call next time?

---

*Sources verified 2026-09-26. Vendor URLs and API surfaces change; re-run §2/§3 before trusting any
date-stamped claim here. This document records **how to fetch**, not the specs themselves.*
