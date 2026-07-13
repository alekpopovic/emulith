# Emulith — Codex GPT-5.6 prompt paket 30–44

Ovaj paket nastavlja prethodne Emulith pakete `00–14` i `15–29`. Njegov cilj je da AWS-first `v0.2.0` projekat proširi u prvi stvarni multi-cloud POC sa Azure Storage podrškom i pripremi `v0.3.0`.

Svaki prompt je zaseban Markdown fajl i predviđen je kao poseban Codex task u istom repozitorijumu. Promptovi su napisani na engleskom zbog preciznijeg rada sa protokolima, SDK-ovima, testovima i acceptance kriterijumima.

## Redosled

| Broj | Fajl | Cilj |
|---:|---|---|
| 30 | `30-azure-provider-gateway-and-endpoints.md` | Azure provider i zasebni Blob/Queue/Table listeneri |
| 31 | `31-azure-storage-account-and-auth-contract.md` | Development account, SharedKey/SAS parser i connection string |
| 32 | `32-blob-container-lifecycle.md` | Blob container lifecycle, metadata i listing |
| 33 | `33-block-blob-core-crud.md` | Osnovni Block Blob upload/download/properties/delete |
| 34 | `34-block-blob-staging-and-sdk-upload.md` | Put Block, Put Block List i high-level SDK upload |
| 35 | `35-blob-listing-ranges-and-conditions.md` | Blob listing, range download i conditional requests |
| 36 | `36-azure-queue-lifecycle.md` | Azure Queue lifecycle, metadata i listing |
| 37 | `37-azure-queue-message-semantics.md` | Message TTL, visibility, pop receipt, update/delete/clear |
| 38 | `38-azure-table-service-and-entity-crud.md` | Table lifecycle, typed entities, CRUD i ETag |
| 39 | `39-azure-table-odata-query-and-pagination.md` | OData parser, filter, select, top i continuation |
| 40 | `40-azure-table-batch-transactions.md` | Multipart entity group transactions i rollback |
| 41 | `41-azure-sdk-compatibility-suite.md` | Formalni official Azure SDK compatibility paket |
| 42 | `42-azure-manifest-compose-and-dev-ux.md` | Manifest resursi, CLI, Compose i Azure demo |
| 43 | `43-azure-state-migration-export-import-and-matrix.md` | Migracije, snapshot i multi-provider compatibility report |
| 44 | `44-v0.3-hardening-and-release.md` | Multicloud hardening i v0.3 GO/NO-GO |

## Način rada

1. Završiti i pregledati prethodne promptove `00–29`.
2. Promptove `30–44` izvršavati strogo redom.
3. Svaki prompt pokrenuti kao poseban Codex task/thread.
4. Ne pokretati dva prompta paralelno nad istim repozitorijumom.
5. Posle svakog zadatka pregledati diff, migracije, SDK testove i compatibility status.
6. Napraviti commit tek nakon ljudskog pregleda završenog koraka.
7. Ne preći na sledeći prompt dok obavezne provere prethodnog nisu zelene ili precizno blokirane okruženjem.

## Ciljani rezultat

Posle prompta 44 Emulith bi trebalo da ima testirani lokalni subset:

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
```

Uz:

```text
četiri lokalna listenera
SQLite/filesystem state
schema migracije
snapshot export/import
Docker
official SDK compatibility report
CI
release artifacts
SBOM/checksums
```

To nije puna Azure ili AWS kompatibilnost, ne sprovodi stvarnu cloud autentikaciju i nije namenjeno produkciji.

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
make docker-build
make release-snapshot
make release-check
```

Docker i release provere zavise od dostupnosti alata u lokalnom okruženju, ali svaki prompt zahteva precizno prijavljivanje ograničenja.
