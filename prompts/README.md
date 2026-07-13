# Emulith — Codex GPT-5.6 prompt paket 15–29

Ovaj paket nastavlja prvi Emulith paket (`00–14`) i vodi projekat od funkcionalnog `v0.1.0-poc` ka stabilnijem AWS-first `v0.2.0` izdanju.

Promptovi su napisani na engleskom zbog preciznijeg izvršavanja tehničkih zahteva u Codex-u. Svaki prompt je zaseban Markdown fajl i treba ga pokretati kao poseban Codex task u istom repozitorijumu.

## Redosled

| Broj | Fajl | Cilj |
|---:|---|---|
| 15 | `15-release-automation-and-artifacts.md` | Release buildovi, checksum, SBOM, signing/provenance i multi-arch image |
| 16 | `16-state-versioning-export-import.md` | Verzionisan state snapshot, bezbedan export/import i migracije |
| 17 | `17-service-registry-and-provider-boundaries.md` | Registry servisa i jasne provider granice |
| 18 | `18-compatibility-matrix-automation.md` | Machine-readable compatibility matrica povezana sa SDK testovima |
| 19 | `19-dynamodb-protocol-and-data-model.md` | DynamoDB AWS JSON protokol, AttributeValue i storage osnova |
| 20 | `20-dynamodb-table-lifecycle.md` | Create/Describe/List/Delete Table |
| 21 | `21-dynamodb-item-crud.md` | Put/Get/Delete/početni UpdateItem |
| 22 | `22-dynamodb-expressions-and-conditions.md` | Lexer/parser/evaluator za izraze i uslovne upise |
| 23 | `23-dynamodb-query-scan-pagination.md` | Query, Scan, projection i prava paginacija |
| 24 | `24-dynamodb-batch-operations.md` | BatchGetItem i BatchWriteItem |
| 25 | `25-dynamodb-sdk-compatibility-suite.md` | Formalni DynamoDB SDK, persistence i concurrency testovi |
| 26 | `26-sns-core-topics-and-publish.md` | SNS topics i Publish |
| 27 | `27-sns-sqs-subscriptions.md` | SNS→SQS subscription i cross-service delivery |
| 28 | `28-cloudwatch-logs-subset.md` | CloudWatch Logs grupe, streamovi, događaji i ograničeni filteri |
| 29 | `29-v0.2-hardening-and-release.md` | Migracije, hardening, acceptance gate i v0.2 GO/NO-GO |

## Preporučeni način rada

1. Završiti i pregledati paket `00–14`.
2. Pokretati promptove `15–29` strogo redom.
3. Svaki prompt izvršiti u posebnom Codex task/thread-u.
4. Ne pokretati dva prompta paralelno nad istim repozitorijumom.
5. Posle svakog zadatka pregledati diff, test rezultate i compatibility status.
6. Commit praviti tek nakon ljudskog pregleda završenog koraka.
7. Ne prelaziti na sledeći prompt dok prethodni nema zelene obavezne testove ili jasno dokumentovano ograničenje okruženja.

## Ciljani rezultat

Posle prompta 29 Emulith bi trebalo da ima testirani lokalni subset:

```text
AWS STS
AWS S3
AWS SQS
AWS DynamoDB
AWS SNS
AWS CloudWatch Logs
```

Uz:

```text
SQLite/filesystem state
schema migrations
snapshot export/import
Docker
SDK compatibility report
CI
release artifacts
SBOM/checksums
```

To i dalje nije puna AWS kompatibilnost niti proizvodni servis.

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
make docker-build
make release-snapshot
make release-check
```

Neke Docker ili release provere zavise od dostupnosti alata u lokalnom okruženju, ali promptovi zahtevaju precizno prijavljivanje svakog takvog ograničenja.
