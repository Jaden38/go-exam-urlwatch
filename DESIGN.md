# DESIGN — URLWatch

## 1. Découpage en packages et frontières d'interface

Les types métier, les interfaces et les règles de validation vivent dans
`internal/domain`, dont tous les autres packages dépendent (et jamais l'inverse).
Les deux frontières d'interface sont `Checker` et `Store`, déclarées dans
`domain` mais **implémentées ailleurs** (`checker.HTTP`, `store.Memory`) : c'est
le consommateur (le domaine et l'API) qui possède le contrat, l'implémentation
s'y conforme. `cmd/urlwatch/main.go` reste mince : il ne fait que construire les
implémentations concrètes et les injecter dans `api.NewServer`. On peut donc
remplacer le Checker réel par un mock ou le Store mémoire par SQLite sans toucher
au reste.

## 2. Modèle de concurrence

`pool.Run` lance exactement `min(concurrency, len(urls))` workers ; le nombre
d'appels HTTP simultanés ne peut donc jamais dépasser la borne demandée. Les
channels `jobs` et `results` sont **non bufferisés** : la distribution sert
directement de backpressure (le fan-out se bloque tant qu'aucun worker n'est
libre), ce qui évite d'accumuler des URLs en attente et garde l'usage mémoire
plat. Inutile de bufferiser pour une charge bornée à 100 URLs.

Échecs partiels : un échec (DNS, timeout, 5xx) n'est pas une erreur du lot, c'est
un `CheckResult` avec `ok:false`. Le lot réussit toujours et le résumé compte
`up`/`down`. **Toutes** les URLs reçoivent un résultat, même après annulation :
le fan-out émet la liste entière, et un context déjà expiré fait retourner
immédiatement un résultat en échec — d'où un résumé toujours cohérent.

Deux niveaux de `context` : un timeout global par lot (calculé d'après le nombre
de vagues `URLs / concurrence` × timeout par URL, plus une marge) et un timeout
par URL dérivé pour chaque appel. L'annulation du client (`r.Context()`) se
propage aux deux.

## 3. Fuites de goroutines

Risque principal : un worker bloqué sur l'envoi dans `results`, ou le fan-out
bloqué sur `jobs`. On l'évite ainsi : la goroutine de lecture consomme `results`
jusqu'à sa fermeture ; `results` est fermé par une goroutine dédiée **après**
`wg.Wait()` ; `jobs` est fermé par le fan-out via `defer`. Chaque worker sort de
sa boucle `range jobs` à la fermeture du channel. Sur annulation, les `Check`
retournent vite (le HTTP respecte le context), donc rien ne reste bloqué. Le test
`TestRunCancellation` échouerait (timeout) en cas de fuite ou de deadlock, et
toute la suite passe sous `-race`.

## 4. Stratégie d'erreurs

- **Sentinelle** `domain.ErrBatchNotFound`, enveloppée par le store
  (`fmt.Errorf("...: %w", err)`) et détectée par `errors.Is` dans le handler →
  `404 batch_not_found`.
- **Type personnalisé** `domain.ValidationError` (porte le champ fautif),
  détecté par `errors.As` → `400 invalid_request` avec un message exploitable.
- La couche API est seule responsable de la traduction erreur → code HTTP ;
  le domaine ne connaît pas le HTTP.

## 5. Pourquoi Go ici

1. **Concurrence native** : le worker pool tient en quelques goroutines + deux
   channels + un `WaitGroup`, sans pool de threads ni boucle d'événements
   externe — exactement le cœur de ce service.
2. **`context` de bout en bout** : annulation et timeout se propagent
   uniformément du handler HTTP jusqu'à `http.Client`, sans plomberie manuelle.
3. **Bibliothèque standard suffisante** : `net/http` (routage 1.22), `log/slog`
   et `encoding/json` couvrent tout le projet — zéro dépendance, build trivial.

**Limite ressentie** : l'absence de génériques sur les helpers JSON et la
verbosité du `if err != nil` répété ; et distinguer « champ absent » de « zéro »
impose des `*int` dans le DTO, là où un type optionnel dédié serait plus clair.

## Extensions non traitées

Persistance SQLite, mode asynchrone (`202 Accepted`) et pagination restent
derrière les interfaces existantes : `store.Memory` peut être remplacé par un
`store.SQLite` sans changer l'API. Priorité donnée à un socle simple, correct et
testé.
