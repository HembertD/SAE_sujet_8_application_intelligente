# Wrapper Git

Fichier source associé : src/git_commit_release_notes_generator/service/git_wrapper.py

## Description

Ce module encapsule les interactions avec GitPython pour récupérer les diffs, filtrer les secrets et fournir un contrat de données exploitable par la génération de messages de commit.

## Responsabilités

- ouvrir un dépôt Git à partir d'un chemin local ;
- récupérer le diff indexé (`git diff --cached`) ;
- normaliser les sections de diff en objets `DiffFile` ;
- détecter les fichiers binaires "opaques" (images, archives, ...) et les exclure des payloads LLM ;
- convertir les formats de document "riches" supportés (`.docx`, `.pdf`) en diff textuel exploitable, par extraction de contenu plutôt que de les traiter comme des binaires ignorés ;
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

Représente un commit de l'historique Git, avec son SHA, message, auteur, date et éventuel tag associé (`tag`).

### `get_recent_commits`

Récupère l'historique complet ou récent des commits avec leurs tags associés pour alimenter le sélecteur interactif TUI.

### Diff textuel des documents riches (`.docx`, `.pdf`)

Pour que l'IA puisse "lire" les changements dans des formats que Git traite nativement comme des binaires opaques, `get_staged_diff()` tente une conversion en diff textuel avant d'exclure un fichier binaire :

- `_extract_docx_text(data)` / `_extract_pdf_text(data)` : extraction de texte brut via `python-docx` (paragraphe par paragraphe) et `pypdf` (page par page). Dépendances optionnelles importées à la demande (coût de chargement non payé si le dépôt n'en contient pas).
- `_DOCUMENT_TEXT_EXTRACTORS` : associe une extension à son extracteur. Seules `.docx` et `.pdf` sont supportées ; les autres binaires (images, archives, ...) gardent le comportement existant (exclus du payload LLM).
- `_read_staged_blob_bytes(file_path)` : lit le contenu indexé (staged) directement depuis l'object database Git (`repo.odb`), sans dépendre du fichier sur disque qui a pu diverger depuis le `git add`.
- `_read_head_blob_bytes(file_path)` : lit la version `HEAD` du fichier, ou `None` s'il est nouveau (pas encore commité).
- `_try_build_document_diff(diff_file)` : orchestre l'extraction des deux côtés (ancien/nouveau, `None` si ajout ou suppression), calcule un diff ligne à ligne (`difflib.unified_diff`) et renvoie un `DiffFile` texte (`is_binary=False`) prêt à suivre le chemin standard (sanitisation, troncature). Renvoie `None` en cas d'extension non supportée ou d'échec d'extraction (document corrompu, dépendance manquante) : l'appelant conserve alors le comportement binaire existant plutôt que de planter.

## Rôle dans le flux

Le wrapper est le point d'entrée entre le dépôt Git et les composants de génération. Il constitue le contrat technique qui alimente le module de génération de commits et de release notes.
