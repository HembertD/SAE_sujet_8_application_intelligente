# Script de génération du PDF de documentation

Fichier source associé : scripts/generate_documentation_pdf.py

## Description

Ce script assemble l'ensemble des pages Markdown référencées dans la navigation de `mkdocs.yml` en un document unique, puis le convertit en PDF via `pandoc` et `xelatex`.

## Flux de travail

1. `collect_pages()` parcourt récursivement la clé `nav` de `mkdocs.yml` (dictionnaires et listes) et collecte les couples `(titre, chemin relatif)` sans doublon.
2. `build_markdown()` lit chaque page (chemin relatif à la racine, avec repli vers `docs/`), préfixe son contenu par un titre `# {titre}`, concatène le tout après une table des matières (`\tableofcontents`), et écrit le résultat dans `.tmp-doc/documentation.md`.
3. `main()` vérifie la présence de `pandoc` et `xelatex` (message d'installation `apt-get` sinon), puis lance `pandoc` avec les options `--from=markdown+raw_tex --standalone --toc --number-sections --pdf-engine=xelatex`, des marges de 1 pouce, des liens colorés en bleu et les polices `DejaVu Sans` / `DejaVu Sans Mono`, vers `documentation.pdf` à la racine.

## Points clés

- Toute page référencée mais introuvable lève `FileNotFoundError` avec le chemin fautif.
- Le script ne dépend que de `pathlib`, `shutil`, `subprocess` et `yaml`.

## Commentaires de fonction (dupliqués du code source)

- `collect_pages()` : Collecte les pages de documentation sans doublon.
- `build_markdown()` : Assemble les pages Markdown en document unique.
- `main()` : Vérifie les outils et génère le PDF de documentation.
