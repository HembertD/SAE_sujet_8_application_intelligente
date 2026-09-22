# Déploiement de la documentation

Cette page explique comment la documentation du projet est construite, publiée et mise à disposition via GitHub Pages.

## Objectif

Le projet utilise MkDocs Material pour générer un site de documentation statique à partir des fichiers Markdown présents dans le dossier `docs/`. Une fois la documentation validée, elle est déployée automatiquement sur GitHub Pages.

## Fichiers concernés

- `mkdocs.yml` : configuration du site MkDocs
- `.github/workflows/documentation.yml` : workflow GitHub Actions de build et déploiement
- `docs/` : contenu Markdown de la documentation

## Configuration MkDocs

Le fichier `mkdocs.yml` définit :

- le nom du site : `SAE - Application intelligente`
- le thème : `material`
- la langue : `fr`
- la navigation du site via la clé `nav`

Le build MkDocs transforme les fichiers Markdown en un site statique dans le dossier `site/`.

## Workflow de déploiement

Le workflow GitHub Actions est défini dans `.github/workflows/documentation.yml`.

### Déclenchement

Le déploiement est exécuté lors de :

- un `push` sur la branche `main`
- un déclenchement manuel via `workflow_dispatch`

### Étapes du job `build`

1. `actions/checkout` : récupération du code source
2. `actions/setup-python` : installation de Python 3.12
3. installation de MkDocs Material
4. exécution de `mkdocs build`
5. `actions/configure-pages` : préparation de GitHub Pages
6. `actions/upload-pages-artifact` : publication de la sortie de build dans un artefact

### Étapes du job `deploy`

Le job `deploy` dépend du job `build` et publie le contenu sur l’environnement GitHub Pages via :

- `actions/deploy-pages@v4`

Cela permet d’exposer la documentation en ligne, selon l’URL définie par le dépôt GitHub Pages.

## Résultat attendu

Après un push sur `main`, GitHub Actions génère le site puis le publie. La documentation devient alors accessible via la page publique GitHub Pages du dépôt.

## Bonnes pratiques

- garder les fichiers Markdown dans `docs/` cohérents avec la structure du projet
- mettre à jour `mkdocs.yml` si un nouveau chapitre ou une nouvelle page est ajouté
- valider le build localement avec `mkdocs build` avant de pousser

## Commande de vérification locale

Pour construire le site en local :

```bash
mkdocs build
```

Cette commande crée le dossier `site/`, qui est le même contenu que celui déployé sur GitHub Pages.
