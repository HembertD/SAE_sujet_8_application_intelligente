# Tests du contexte de merge

Fichier source associé : test/test_git_wrapper_merge.py

## Description

Ces tests vérifient le comportement de `GitWrapper` sur de vrais dépôts Git
temporaires, sans réseau ni fixture externe. Ils couvrent le diff indexé, les
conflits de merge, les opérations Git en cours et le payload envoyé au LLM.

## Ce qui est testé

- l'état Git (`merge`, `rebase`, `cherry-pick`) et la présence de conflits ;
- l'extraction des stages Git (`base`, `ours`, `theirs`) ;
- le parsing des marqueurs de conflit, y compris le style `diff3` ;
- les conflits ajout/ajout, suppression/modification et binaires ;
- l'exclusion des fichiers non fusionnés du diff indexé classique ;
- le merge propre en cours et le conflit résolu puis indexé ;
- le payload `get_merge_payload()` / `get_llm_payload()` ;
- la troncature par fichier et le budget global du payload.

## Cas de test principaux

### 1. Conflit de merge classique

Le test crée deux branches qui modifient la même ligne, provoque un conflit,
puis vérifie :

- `operation == "merge"`, `in_progress == True`, `has_conflicts == True` ;
- `merge_head`, `merge_base` et `merge_heads` sont renseignés ;
- `get_staged_diff()` ne retourne pas le fichier non fusionné ;
- `get_conflict_files()` expose le type `both_modified`, les contenus des
  stages et une section avec `ours`, `theirs` et les numéros de ligne ;
- le payload contient `repository`, `summary`, `files` et `conflicts`.

### 2. Style `diff3`

Avec `merge.conflictStyle=diff3`, la section extraite contient aussi la version
`base` issue du marqueur `|||||||`.

### 3. Types de conflits

Les scénarios ajout/ajout et suppression/modification vérifient les types
`both_added`, `deleted_by_us` et la présence ou l'absence de `base_content`.
Un conflit binaire est conservé en métadonnées avec
`resolution_supported=False` et aucune section.

### 4. Merge propre et conflit résolu

Un merge sans conflit avec `--no-commit` expose les fichiers entrants dans le
diff indexé. Après résolution manuelle et `git add`, le fichier n'est plus en
conflit mais reste indexé : le payload le place dans `files`, pas dans
`conflicts`.

### 5. Opérations Git

Les conflits de `cherry-pick` et de `rebase` sont détectés via les fichiers
internes de Git (`CHERRY_PICK_HEAD`, `rebase-merge`, etc.).

### 6. Troncature

Un conflit volumineux est tronqué par `max_chars`; le budget global
`total_max_chars` réduit ensuite le payload et renseigne `summary.truncated`
ainsi que `summary.truncated_files`.

## Points de vigilance

- Les tests utilisent `git init -b main` et configurent l'identité localement.
- Les dépôts sont supprimés automatiquement à la fin de chaque test.
- Les conflits binaires ne doivent jamais être proposés à une résolution
  automatique.
- Les clés historiques `summary` et `files` doivent rester présentes dans le
  payload.
