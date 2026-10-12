# Wrapper Git

Fichier source associé : src/git_commit_release_notes_generator/service/git_wrapper.py

## Description

Ce module encapsule les interactions avec GitPython pour récupérer les diffs, filtrer les secrets et fournir un contrat de données exploitable par la génération de messages de commit.

## Responsabilités

- ouvrir un dépôt Git à partir d'un chemin local ;
- récupérer le diff indexé (`git diff --cached`) ;
- normaliser les sections de diff en objets `DiffFile` ;
- détecter les fichiers binaires et ignorer les diff binaires dans les payloads LLM ;
- masquer les secrets dans le texte avant transmission à l'IA ;
- exposer des payloads structurés pour le reste de l'application.

## Points clés

### `GitWrapper`

La classe centrale s'occupe de la lecture du dépôt et du parsing des diff Git. Elle renvoie une liste de `DiffFile`, chaque objet contenant :

- `path` : chemin du fichier modifié ;
- `status` : type de changement (`A`, `M`, `D`, `R`, ... ) ;
- `added` / `removed` : nombre de lignes modifiées ;
- `is_binary` : indicateur de fichier binaire ;
- `patch` : contenu brut du diff.

### `GitWrapperError`

Exception métier utilisée pour signaler un échec de lecture Git ou un problème d'extraction du diff.

### `CommitInfo`

Représente un commit de l'historique Git, avec son SHA, message, auteur et date.

## Rôle dans le flux

Le wrapper est le point d'entrée entre le dépôt Git et les composants de génération. Il constitue le contrat technique qui alimente ensuite le module de génération de messages via les objets `DiffFile`.

## Détails déplacés du code source

- `SECRET_PATTERNS` : trois expressions régulières (clés `api[_-]?key|secret|token|password|auth` avec valeur d'au moins 8 caractères, tokens `ghp_` de 36 caractères, JWT `ey...`). `_sanitize_diff_text()` les remplace par `[REDACTED_SECRET]` en conservant le nom du champ capturé quand il existe ; `sanitize_diff()` en est le point d'entrée public.
- `_parse_diff()` : normalise les lignes `diff --git`, découpe par fichier, résout `a/...`/`b/...` et `/dev/null`. Le statut vaut `A` (marqueur `new file mode` ou ancien chemin `/dev/null`), `D` (marqueur `deleted file mode` ou nouveau chemin `/dev/null`), `R` (renommage), sinon `M`. Les marqueurs de mode ne sont matchés qu'en début de ligne pour ne pas confondre le contenu ajouté/supprimé. Un fichier est binaire si sa section contient `Binary files` ou `GIT binary patch` (compteurs à 0) ; sinon `added`/`removed` comptent les lignes `+`/`-` hors en-têtes `+++`/`---`.
- `get_staged_diff(max_chars=8000)` : lit `git diff --cached --no-color --diff-filter=ACMRTUXBD`, écarte les fichiers binaires, et tronque tout `patch` dépassant `max_chars` avec le marqueur `... [DIFF TRUNCATED: TROP VOLUMINEUX] ...`.
- `get_staged_diff_as_json()` : liste prête pour JSON via `to_dict()` ; `get_staged_diff_payload()` / `get_llm_payload()` ajoutent un résumé (`file_count`, `binary_file_count`, `truncated`, sans bloc prompt interne).
- `get_commits_between_tags(tag_a, tag_b)` : parcourt `tag_a..tag_b` et renvoie des `CommitInfo` (SHA court à 7 caractères, message nettoyé, auteur, date ISO).
- `commit(message)` : refuse si rien n'est indexé (`is_dirty(index=True)` faux) et renvoie le SHA court du commit créé.
