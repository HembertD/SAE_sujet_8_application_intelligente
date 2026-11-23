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
- détecter l'opération Git en cours (`merge`, `rebase`, `cherry-pick`, `revert`) ;
- exposer les fichiers non fusionnés et leurs sections de conflit ;
- produire un payload unique diff + état Git + conflits pour le LLM.

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

Représente un commit de l'historique Git, avec son SHA, message, auteur, date et éventuel tag associé (`tag`).

### `get_recent_commits`

Récupère l'historique complet ou récent des commits avec leurs tags associés pour alimenter le sélecteur interactif TUI.

## Contexte de merge et payload LLM

### `get_git_state()`

Retourne un objet `GitState` décrivant le contexte courant :

- `branch` : branche active, ou `HEAD` si le dépôt est détaché ;
- `head` : SHA court du commit courant ;
- `operation` : `none`, `merge`, `rebase`, `cherry-pick` ou `revert` ;
- `in_progress` : indique qu'une opération Git est en cours ;
- `has_conflicts` : indique la présence de chemins non fusionnés ;
- `merge_head` / `merge_heads` : SHA court(s) de la ou des branches entrantes ;
- `merge_base` : ancêtre commun avec `MERGE_HEAD`, quand il est calculable ;
- `head_parents` : parents du commit courant ;
- `is_merge_commit` : vrai si le commit courant a plusieurs parents.

### `get_conflict_files()` et `get_conflict_sections()`

`get_conflict_files()` lit les stages Git (`:1:` base, `:2:` ours, `:3:` theirs)
et le fichier de travail :

- le type de conflit est déduit des stages (`both_modified`, `both_added`,
  `deleted_by_us`, `deleted_by_them`, etc.) ;
- les versions complètes `base_content`, `ours_content` et `theirs_content`
  servent de repli quand les marqueurs ne sont pas exploitables ;
- les sections sont extraites des marqueurs `<<<<<<<`, `|||||||`, `=======`,
  `>>>>>>>`, avec les numéros de ligne et un contexte limité ;
- les conflits binaires sont conservés en métadonnées (`is_binary=True`,
  `resolution_supported=False`, `sections=[]`) ;
- les contenus trop volumineux sont tronqués et signalés par `truncated`.

`get_conflict_sections()` aplatit les sections de tous les fichiers en conflit.

### `get_merge_payload()` et `get_llm_payload()`

`get_merge_payload()` construit le payload complet. `get_llm_payload()` en est
l'alias public destiné au modèle. Il conserve les clés historiques `summary`
et `files`, et ajoute `repository` et `conflicts` :

```json
{
  "repository": {
    "branch": "main",
    "head": "abc1234",
    "operation": "merge",
    "in_progress": true,
    "has_conflicts": true,
    "merge_head": "def5678",
    "merge_heads": ["def5678"],
    "merge_base": "0123456",
    "head_parents": ["0123456"],
    "is_merge_commit": false
  },
  "summary": {
    "file_count": 0,
    "conflict_count": 1,
    "binary_file_count": 0,
    "binary_conflict_count": 0,
    "truncated": false,
    "truncated_files": []
  },
  "files": [],
  "conflicts": [
    {
      "path": "src/app.py",
      "conflict_type": "both_modified",
      "resolution_supported": true,
      "is_binary": false,
      "reason": "",
      "truncated": false,
      "base_content": "...",
      "ours_content": "...",
      "theirs_content": "...",
      "sections": [
        {
          "file_path": "src/app.py",
          "section_index": 0,
          "start_line": 12,
          "end_line": 20,
          "base": "...",
          "ours": "...",
          "theirs": "...",
          "context_before": "...",
          "context_after": "...",
          "conflict_type": "both_modified",
          "resolution_supported": true,
          "source": "markers"
        }
      ]
    }
  ]
}
```

Le payload est borné par deux paramètres :

- `max_chars` : taille maximale d'un patch ou d'un contenu de conflit ;
- `total_max_chars` : budget global du payload, les conflits étant conservés en
  priorité ;
- `context_lines` : nombre de lignes de contexte autour d'une section.

`get_staged_diff()` exclut volontairement les chemins non fusionnés (`U`) :
les conflits sont exposés séparément pour ne pas être confondus avec un diff
indexé classique.

## Rôle dans le flux

Le wrapper est le point d'entrée entre le dépôt Git et les composants de génération. Il constitue le contrat technique qui alimente le module de génération de commits et de release notes.
