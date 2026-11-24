# Tests du wrapper Git

Fichier source associé : test/test_git_wrapper.py

## Description

Ce module valide le comportement du wrapper Git utilisé pour analyser les diffs et préparer les données envoyées au système de génération de commit.

## Ce qui est testé

- le parsing d’un diff Git en fichiers structurés ;
- la détection des modifications par fichier ;
- la suppression des données sensibles dans les patches ;
- la création d’un payload JSON exploitable par le modèle ou par l’orchestrateur applicatif ;
- l'extraction de texte depuis des documents `.docx`/`.pdf` et leur conversion en diff textuel exploitable ;
- la détection des images modifiées (ajout/modification/suppression) avec leurs octets avant/après, sans appel réseau.

## Cas de test principaux

### 1. Parsing d’un diff multi-fichiers

Le test vérifie que le diff est découpé correctement en plusieurs objets, avec pour chaque fichier :

- le chemin du fichier ;
- le statut de modification ;
- les lignes ajoutées et supprimées ;
- le patch complet associé.

### 2. Sanitisation des secrets

Le test simule un contenu contenant une clé ou une valeur sensible et vérifie que le mécanisme de redaction remplace bien les informations sensibles par une valeur de type `[REDACTED_SECRET]`.

### 3. Structure du payload JSON

Le test s’assure que les objets renvoyés par le wrapper sont convertibles en payloads normalisés, sans informations parasites ou inutiles. Il vérifie notamment la présence des champs attendus comme `path`, `status`, `patch` et `is_binary`.

### 4. Extraction de texte des documents riches

- `test_extract_docx_text_reads_paragraphs` / `test_extract_docx_text_reads_table_cells` / `test_extract_pdf_text_reads_content` : construisent un `.docx` (via `python-docx`, paragraphes et tableaux) et un `.pdf` minimal mais valide (flux de contenu PDF écrit à la main, lu par `pypdf`) directement en mémoire, et vérifient que `_extract_docx_text`/`_extract_pdf_text` en retrouvent le texte — y compris le contenu des cellules de tableau, ignoré par défaut par `python-docx.paragraphs`.
- `test_get_staged_diff_converts_modified_docx_to_text_diff` : sur un vrai dépôt Git temporaire, modifie un `.docx` entre deux commits/staging et vérifie que `get_staged_diff()` renvoie un `DiffFile` texte (`is_binary=False`) dont le patch contient l'ancien et le nouveau contenu, avec des compteurs `added`/`removed` cohérents.
- `test_get_staged_diff_converts_newly_added_pdf_to_text_diff` : même principe pour un `.pdf` nouvellement ajouté (`status="A"`, pas de contenu "ancien"). Vérifie aussi explicitement l'absence de l'en-tête brut `%PDF` dans le patch et la sérialisation JSON sans erreur (`json.dumps(entry.to_dict(), ensure_ascii=False)`) — régression du bug corrigé où un PDF que Git classe comme "texte" (`is_binary=False`) laissait passer son bytecode brut non-UTF8 dans le patch, provoquant un `UnicodeEncodeError` à la sérialisation.
- `test_get_staged_diff_handles_corrupted_document_without_crashing` : un `.docx`/`.pdf` corrompu (octets arbitraires) est exclu silencieusement du diff, sans exception.
- `test_get_staged_diff_handles_deleted_document` : la suppression d'un `.docx` (`status="D"`) produit un diff textuel exploitable et sérialisable en JSON.
- `test_get_staged_diff_still_excludes_unsupported_binary_files` : vérifie la non-régression du comportement existant — un binaire sans extracteur dédié (ex: `.png`) reste exclu du diff exploitable par le LLM.

### 5. Détection des images modifiées (`get_staged_image_changes`)

Sur de vrais dépôts Git temporaires (PNG minimaux mais valides, octets distincts entre versions) :

- `test_get_staged_image_changes_detects_addition` : `status="A"`, `new_bytes` renseigné, `old_bytes=None`.
- `test_get_staged_image_changes_detects_modification` : `status="M"`, `old_bytes`/`new_bytes` correspondent bien aux deux versions committées/stagées.
- `test_get_staged_image_changes_detects_deletion` : `status="D"`, `old_bytes` renseigné (contenu de la version supprimée), `new_bytes=None`.
- `test_get_staged_image_changes_ignores_text_and_document_files` : un `.txt` et un `.docx` stagés ne produisent aucune `ImageChange` (périmètre strictement limité aux extensions image).

## Points de vigilance

- La logique de parsing doit rester tolérante aux variations de format de diff Git.
- La redaction des secrets doit être systématique avant toute transmission hors du contexte local.
- Les payloads doivent rester structurés et légers pour éviter d’encombrer le modèle LLM.
- L'extraction de documents riches doit échouer silencieusement (repli sur le comportement binaire existant) plutôt que de faire planter le diff sur un document corrompu ou un format inattendu.
- L'extraction doit être tentée pour toute extension supportée **indépendamment** du flag `is_binary` posé par Git (un `.pdf` sans octet nul dans ses premiers kilo-octets peut être classé "texte" par Git) : ne jamais se fier uniquement à `diff_file.is_binary` pour décider de convertir un document, sous peine de laisser fuiter du bytecode brut non-UTF8 dans le patch.
