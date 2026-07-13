# Emulith — Codex GPT-5.6 prompt paket 60–74

Ovaj paket nastavlja Emulith promptove `00–59`. Cilj je da tro-cloud `v0.4.0` projekat proširi lokalnim serverless runtime-om i pripremi `v0.5.0`.

Svaki prompt je zaseban Markdown fajl i predviđen je kao poseban Codex task u istom repozitorijumu. Promptovi su napisani na engleskom zbog preciznijeg rada sa OCI/Docker runtime-om, provider protokolima, event envelope-ima, retry semantikom i acceptance kriterijumima.

## Redosled

| Broj | Fajl | Cilj |
|---:|---|---|
| 60 | `60-serverless-execution-engine-contract.md` | Interni function/revision/invocation/runtime ugovori i state machine |
| 61 | `61-oci-container-runtime-and-sandbox.md` | Docker/OCI execution backend, warm pool i bezbednosne granice |
| 62 | `62-function-artifacts-build-cache-and-manifest.md` | Function manifest, image/source build, revision i cache |
| 63 | `63-durable-invocation-queue-retries-and-dlq.md` | Durable event delivery, retry, concurrency, scheduling i DLQ |
| 64 | `64-aws-lambda-control-plane.md` | AWS Lambda image-based control-plane subset |
| 65 | `65-aws-lambda-runtime-api-and-invoke.md` | Lambda Runtime API, Invoke, cold/warm i async delivery |
| 66 | `66-aws-lambda-event-sources-eventbridge-and-scheduler.md` | SQS/SNS/S3 triggeri, EventBridge i scheduler |
| 67 | `67-azure-functions-custom-handler-http-and-timer.md` | Azure custom handler, HTTP i timer trigger |
| 68 | `68-azure-functions-bindings-and-event-grid.md` | Azure Queue/Blob binding i Event Grid subset |
| 69 | `69-gcp-functions-framework-and-cloud-run-functions.md` | GCP HTTP/CloudEvent Functions Framework runtime |
| 70 | `70-gcp-eventarc-cloud-scheduler-and-service-triggers.md` | Pub/Sub/Storage/Firestore triggeri, Eventarc i Scheduler |
| 71 | `71-serverless-observability-logs-traces-and-metrics.md` | Logs, invocation history, W3C tracing, OTLP i metrics |
| 72 | `72-serverless-compatibility-chaos-and-e2e-suite.md` | Formalni compatibility, chaos i multicloud E2E testovi |
| 73 | `73-serverless-manifest-docker-state-and-dev-ux.md` | Kompletan manifest/CLI/Compose/snapshot/demo UX |
| 74 | `74-v0.5-serverless-hardening-and-release.md` | Security/correctness hardening i v0.5 GO/NO-GO |

## Način rada

1. Završiti i pregledati promptove `00–59`.
2. Izvršavati `60–74` strogo redom.
3. Svaki prompt pokrenuti kao zaseban Codex task/thread.
4. Ne pokretati paralelne taskove nad istim repozitorijumom.
5. Posle svakog koraka pregledati diff, migracije, runtime security, official-client testove i compatibility status.
6. Commit praviti tek posle ljudskog pregleda.
7. Ne prelaziti na sledeći prompt dok obavezne provere prethodnog nisu zelene ili precizno blokirane okruženjem.

## Ciljani rezultat

Posle prompta 74 Emulith bi trebalo da ima lokalni, testirani subset:

```text
Shared:
  OCI/Docker execution engine
  immutable function revisions
  warm container pool
  durable invocation queue
  retry/backoff
  DLQ
  schedules
  logs/traces/metrics

AWS:
  Lambda image custom runtime
  Runtime API
  Invoke
  SQS/SNS/S3 triggers
  EventBridge
  Scheduler

Azure:
  Functions custom handler
  HTTP/Timer
  Queue/Blob bindings
  Event Grid subset

GCP:
  Functions Framework HTTP/CloudEvent
  Pub/Sub/Storage/Firestore triggers
  Eventarc subset
  Cloud Scheduler subset
```

To nije puna kompatibilnost sa managed function platformama i nije jaka multi-tenant izolacija. Docker daemon je eksplicitna lokalna trust granica.

## Završni acceptance gate

```bash
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility
make compatibility-report
make compatibility-check
make demo
make demo-azure
make demo-gcp
make demo-multicloud
make demo-serverless
make demo-multicloud-serverless
make docker-build
make release-snapshot
make release-check
```

Docker i release provere zavise od dostupnosti alata u lokalnom okruženju, ali svaki prompt zahteva precizno prijavljivanje ograničenja.
