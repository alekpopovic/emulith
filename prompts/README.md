# Emulith — Codex GPT-5.6 prompt paket 45–59

Ovaj paket nastavlja Emulith promptove `00–44`. Cilj je da multi-cloud `v0.3.0` projekat proširi GCP servisima i pripremi prvi AWS + Azure + GCP `v0.4.0`.

Svaki prompt je zaseban Markdown fajl i predviđen je kao poseban Codex task u istom repozitorijumu. Promptovi su napisani na engleskom zbog preciznijeg rada sa gRPC/HTTP protokolima, zvaničnim SDK klijentima, migracijama i acceptance kriterijumima.

## Redosled

| Broj | Fajl | Cilj |
|---:|---|---|
| 45 | `45-gcp-provider-grpc-and-endpoints.md` | GCP provider, gRPC osnova i Storage HTTP listener |
| 46 | `46-gcp-project-credentials-and-emulator-env.md` | Lokalni project, resource names, permissive auth i `gcp env` |
| 47 | `47-pubsub-topic-and-subscription-lifecycle.md` | Pub/Sub topics i pull subscriptions |
| 48 | `48-pubsub-publish-pull-and-ack.md` | Publish, Pull, Acknowledge i ack deadline |
| 49 | `49-pubsub-streaming-pull-redelivery-and-ordering.md` | StreamingPull, backpressure, redelivery i ordering keys |
| 50 | `50-pubsub-sdk-compatibility-suite.md` | Formalni official-client Pub/Sub paket |
| 51 | `51-gcs-bucket-lifecycle.md` | Cloud Storage bucket CRUD/listing/metadata |
| 52 | `52-gcs-object-core-crud.md` | Object upload/download/attrs/patch/delete |
| 53 | `53-gcs-resumable-and-multipart-uploads.md` | Multipart i resumable upload protokol |
| 54 | `54-gcs-listing-ranges-generations-and-conditions.md` | Listing, ranges, generation i preconditions |
| 55 | `55-firestore-protocol-value-model-and-persistence.md` | Firestore gRPC, Value model, paths i storage |
| 56 | `56-firestore-document-crud-batches-and-transactions.md` | CRUD, batches, transforms i transactions |
| 57 | `57-firestore-structured-queries-and-sdk-compatibility.md` | RunQuery i formalni Firestore SDK testovi |
| 58 | `58-gcp-manifest-docker-state-and-compatibility-integration.md` | Manifest, Docker, migracije, snapshot i GCP demo |
| 59 | `59-v0.4-hardening-and-release.md` | Three-cloud hardening i v0.4 GO/NO-GO |

## Način rada

1. Završiti i pregledati promptove `00–44`.
2. Izvršavati `45–59` strogo redom.
3. Svaki prompt pokrenuti kao zaseban Codex task/thread.
4. Ne pokretati paralelne taskove nad istim repozitorijumom.
5. Posle svakog koraka pregledati diff, migracije, official-client testove i compatibility status.
6. Commit praviti tek posle ljudskog pregleda.
7. Ne prelaziti na sledeći prompt dok obavezne provere prethodnog nisu zelene ili precizno blokirane okruženjem.

## Ciljani rezultat

Posle prompta 59 Emulith bi trebalo da ima testirani lokalni subset:

```text
AWS:
  STS
  S3
  SQS
  DynamoDB
  SNS
  CloudWatch Logs

Azure:
  Blob Storage
  Queue Storage
  Table Storage

GCP:
  Pub/Sub
  Cloud Storage
  Firestore
```

Uz:

```text
sedam lokalnih listenera
SQLite/filesystem state
schema migracije
snapshot export/import
Docker
official SDK compatibility report
CI
release artifacts
SBOM/checksums
```

To nije puna cloud kompatibilnost, ne sprovodi stvarni IAM/auth i nije namenjeno produkciji.

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
make docker-build
make release-snapshot
make release-check
```

Docker i release provere zavise od dostupnosti alata u lokalnom okruženju, ali svaki prompt zahteva precizno prijavljivanje ograničenja.
