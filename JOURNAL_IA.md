# JOURNAL_IA

## Comment l'IA a été utilisée

J'ai gardé la main sur l'ensemble du code. L'IA n'est intervenue que sur trois
points précis :

- **Orientations** : demander des directions sur l'approche à suivre avant
  d'écrire moi-même le code.
- **Correction des parties les plus complexes** : faire relire et corriger, si
  besoin, ce que j'avais écrit sur les points délicats. Deux exemples :
  - le cœur concurrent `pool.Run` (worker pool + fan-out/fan-in), en particulier
    la fermeture des channels et l'absence de fuite de goroutine ;
  - la propagation des `context` (global vs par URL) et la traduction des
    erreurs métier en codes HTTP via `errors.Is`/`errors.As`.
- **Reformulation** : réécriture de `DESIGN.md` et des commentaires pour qu'ils
  respectent les conventions GoDoc.

Aucun autre usage : l'architecture, le découpage en packages et les choix de
conception sont les miens.