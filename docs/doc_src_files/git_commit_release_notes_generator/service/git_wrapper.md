# Wrapper Git

Fichier source associé : src/git_commit_release_notes_generator/service/git_wrapper.py

## Description

Ce module encapsule les interactions avec GitPython pour récupérer les diffs, filtrer les secrets et fournir un contrat de données exploitable par la génération de messages de commit.

## Responsabilités

- ouvrir un dépôt Git à partir d'un chemin local ;
- récupérer le diff indexé (`git diff --cached`) ;
- normaliser les sections de diff en objets `DiffFile` ;
- détecter les fichiers binaires "opaques" (archives, ...) et les exclure des payloads LLM ;
- convertir les formats de document "riches" supportés (`.docx`, `.pdf`) en diff textuel exploitable, par extraction de contenu plutôt que de les traiter comme des binaires ignorés ;
- exposer les octets bruts des images modifiées (`get_staged_image_changes()`), sans appel réseau, pour une description optionnelle par un modèle de vision (module C, `ollama_client.vision`) ;
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

- `_extract_docx_text(data)` : extraction de texte brut via `python-docx`, paragraphes **et tableaux** (`document.tables`, ignorés par défaut par `document.paragraphs`).
- `_extract_pdf_text(data)` : extraction page par page via `pypdf`. Dépendances optionnelles importées à la demande (coût de chargement non payé si le dépôt n'en contient pas).
- `_DOCUMENT_TEXT_EXTRACTORS` : associe une extension à son extracteur. Seules `.docx` et `.pdf` sont supportées ; les autres binaires (images, archives, ...) gardent le comportement existant (exclus du payload LLM).
- `_read_staged_blob_bytes(file_path)` : lit le contenu indexé (staged) directement depuis l'object database Git (`repo.odb`), sans dépendre du fichier sur disque qui a pu diverger depuis le `git add`.
- `_read_head_blob_bytes(file_path)` : lit la version `HEAD` du fichier, ou `None` s'il est nouveau (pas encore commité).
- `_try_build_document_diff(diff_file, old_path=None)` : orchestre l'extraction des deux côtés (ancien/nouveau, `None` si ajout ou suppression), calcule un diff ligne à ligne (`difflib.unified_diff`) et renvoie un `DiffFile` texte (`is_binary=False`) prêt à suivre le chemin standard (sanitisation, troncature). Renvoie `None` en cas d'extension non supportée ou d'échec d'extraction (document corrompu, dépendance manquante) : l'appelant conserve alors le comportement binaire existant plutôt que de planter. `old_path` sert uniquement à lire la version HEAD depuis un chemin différent de `diff_file.path` (cas du renommage, ci-dessous) — il n'est jamais exposé dans le `DiffFile` renvoyé, qui garde le même contrat qu'avant.
- **Point de vigilance (corrigé)** : un `.pdf` n'a pas forcément d'octet nul dans ses premiers kilo-octets — Git peut donc le classer comme "texte" (`is_binary=False`) et produire un diff ligne à ligne du bytecode PDF brut, qui contient des octets non-UTF8 (surrogates une fois décodés par GitPython). `get_staged_diff()` tente donc l'extraction pour **toute** extension présente dans `_DOCUMENT_TEXT_EXTRACTORS`, que Git l'ait étiqueté binaire ou non — jamais seulement quand `diff_file.is_binary` est vrai. Si l'extraction échoue malgré tout, le fichier est exclu plutôt que de laisser passer ce bytecode brut, pour éviter un crash `UnicodeEncodeError` lors de la sérialisation JSON en aval.
- `_get_staged_rename_sources()` : pour un document renommé (`git mv`), `diff_file.path` porte déjà le nouveau nom — introuvable dans `HEAD`. Cette méthode associe nouveau → ancien chemin via `git diff --cached -M --diff-filter=R --name-status` (détection de renommage explicite, pas activée par défaut par `get_staged_diff()` pour ne pas payer ce calcul sur l'essentiel des diffs). Appelée à la demande, une seule fois par appel à `get_staged_diff()`, uniquement s'il y a au moins un document renommé à traiter. Best-effort : un échec renvoie un dict vide plutôt que de faire planter le diff — le fichier renommé est alors traité comme un ajout complet (comportement antérieur), pas une erreur.

### Images modifiées (`get_staged_image_changes`)

Contrairement aux documents riches ci-dessus, aucune extraction de contenu n'est tentée ici : `get_staged_image_changes()` liste simplement les fichiers image (`.png`, `.jpg`, `.jpeg`, `.gif`, `.bmp`, `.webp`) ajoutés/modifiés/supprimés, avec leurs octets avant/après (réutilise `_read_staged_blob_bytes`/`_read_head_blob_bytes`), sous forme d'`ImageChange(path, status, old_bytes, new_bytes)`. N'effectue aucun appel réseau : c'est à l'appelant (`action_generate_commit`, et uniquement au moment de générer un message de commit) de décider s'il invoque un modèle de vision sur ces octets — `get_staged_diff()` reste donc rapide et sans réseau, inchangé pour les images.

## Rôle dans le flux

Le wrapper est le point d'entrée entre le dépôt Git et les composants de génération. Il constitue le contrat technique qui alimente le module de génération de commits et de release notes.
