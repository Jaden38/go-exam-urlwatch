# URLWatch

Microservice de vérification d'URLs en masse, écrit en Go pour l'examen final du
cours Langage Go. Un client envoie une liste d'URLs ; le service les interroge en
parallèle (code HTTP, latence, disponibilité), agrège les résultats et les expose
via une API REST. Chaque lot est conservé et relisible par son identifiant.

## Prérequis

- Go 1.22 ou supérieur (`log/slog`, routage par méthode du `ServeMux`).
- Aucune dépendance externe : tout repose sur la bibliothèque standard.

## Build, run, test

```bash
go build ./...
go vet ./...
go test ./...          # ajouter -race pour la détection de data races
go run ./cmd/urlwatch   # démarre le serveur sur :8080
```

Variables d'environnement :

| Variable    | Défaut  | Rôle                                            |
|-------------|---------|-------------------------------------------------|
| `ADDR`      | `:8080` | Adresse d'écoute du serveur HTTP.               |
| `LOG_LEVEL` | `info`  | Niveau des logs : `debug`, `info`, `warn`, `error`. |

## API

| Méthode | Chemin             | Rôle                                             |
|---------|--------------------|--------------------------------------------------|
| `POST`  | `/v1/checks`       | Crée et exécute un lot, le persiste, le renvoie. |
| `GET`   | `/v1/checks/{id}`  | Renvoie un lot existant (ou `404`).              |
| `GET`   | `/healthz`         | Sonde de vivacité.                               |

### Créer un lot

```bash
curl -s -X POST localhost:8080/v1/checks \
  -H 'Content-Type: application/json' \
  -d '{
        "urls": ["https://go.dev", "https://exemple.invalid"],
        "options": { "concurrency": 4, "timeout_ms": 2000 }
      }'
```

Réponse `201 Created` :

```json
{
  "batch_id": "b_4f3c1a",
  "created_at": "2026-06-19T14:00:00Z",
  "summary": { "total": 2, "up": 1, "down": 1, "duration_ms": 812 },
  "results": [
    { "url": "https://go.dev", "status_code": 200, "ok": true, "latency_ms": 120 },
    { "url": "https://exemple.invalid", "ok": false, "error": "timeout", "latency_ms": 2001 }
  ]
}
```

Options (toutes optionnelles) : `concurrency` (défaut `8`, borné `1..50`) et
`timeout_ms`, le timeout par URL (défaut `5000`, borné `100..30000`). La liste
`urls` est obligatoire (1 à 100 URLs `http`/`https` valides).

### Relire un lot

```bash
curl -s localhost:8080/v1/checks/b_4f3c1a
```

### Sonde de vivacité

```bash
curl -s localhost:8080/healthz   # {"status":"ok"}
```

### Contrat d'erreur

Toute erreur renvoie le même corps, avec le code HTTP adapté :

```json
{ "error": { "code": "batch_not_found", "message": "aucun lot avec l'id b_x" } }
```

Codes : `400 invalid_request`, `404 batch_not_found`, `405` (méthode non
autorisée, géré par le `ServeMux`), `500 internal`.

## Organisation

```text
cmd/urlwatch/   point d'entrée : câblage, logger, arrêt gracieux
internal/
  domain/       types, interfaces (Checker, Store), erreurs, validation
  checker/      Checker HTTP réel + mock déterministe
  pool/         worker pool borné (fan-out / fan-in, context)
  store/        Store en mémoire
  api/          handlers, routage, middlewares, DTO JSON
```

Voir `DESIGN.md` pour la justification des choix d'architecture.
