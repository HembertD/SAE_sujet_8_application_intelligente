# Déploiement de la documentation

Cette page explique comment la documentation du projet est construite, publiée et mise à disposition via GitHub Pages.

## Objectif

Le projet utilise MkDocs Material pour générer un site de documentation statique à partir des fichiers Markdown présents dans le dossier `docs/`. Une fois la documentation validée, elle est déployée automatiquement sur GitHub Pages.

## Fichiers concernés

- `mkdocs.yml` : configuration du site MkDocs
- `.github/workflows/ci.yml` (jobs `documentation-build` et `deploy-docs`) : workflow GitHub Actions de build et déploiement
- `docs/` : contenu Markdown de la documentation

## Configuration MkDocs

Le fichier `mkdocs.yml` définit :

- le nom du site : `SAE - Application intelligente`
- le thème : `material`
- la langue : `fr`
- la navigation du site via la clé `nav`

Le build MkDocs transforme les fichiers Markdown en un site statique dans le dossier `site/`.

## Prérequis : activation manuelle de GitHub Pages (une seule fois)

!!! warning "À faire par un administrateur du dépôt"
	Avant le premier déploiement, GitHub Pages doit être activé manuellement, car le `GITHUB_TOKEN` du workflow n'a pas le droit de créer le site lui-même (erreur `Resource not accessible by integration` dans `actions/configure-pages` sinon) :

	1. Aller sur **Settings > Pages** du dépôt.
	2. Dans **Build and deployment**, mettre **Source = `GitHub Actions`**.
	3. Relancer le workflow `CI` (automatique au prochain push sur `main`, ou manuel via **Actions > CI > Run workflow** sur `main`).

	Le site est ensuite disponible à l'adresse configurée dans `mkdocs.yml` (`site_url`), soit <https://hembertd.github.io/SAE_sujet_8_application_intelligente/>.

## Prérequis du PDF versionné

!!! warning "Important"
	Le hook `pre-commit` régénère `documentation.pdf` lorsqu'une page de documentation change. Pandoc et XeLaTeX doivent être installés, sinon le commit est bloqué.

	```bash
	sudo apt-get update && sudo apt-get install -y pandoc texlive-xetex texlive-latex-recommended texlive-fonts-recommended lmodern
	```

## Workflow de déploiement

Le workflow GitHub Actions est défini dans `.github/workflows/ci.yml` (jobs `documentation-build` puis `deploy-docs`).

### Déclenchement

Le build de documentation est exécuté lors de :

- un `push` sur toute branche (`branches: ['**']`)
- un `pull_request` ciblant `main` ou `dev`
- un déclenchement manuel via `workflow_dispatch`

Le déploiement officiel sur GitHub Pages (`deploy-docs` via `actions/deploy-pages@v4`) reste réservé à la branche `main`.

### Marqueur de version de développement

Quand la documentation est générée depuis une branche autre que `main`, le titre du site reçoit un suffixe `*` pour signaler qu’il s’agit d’une version en cours de développement.

- `*` = documentation en cours de développement
- sans `*` = documentation officielle de la branche principale

### Étapes du job `build`

1. `actions/checkout` : récupération du code source
2. `actions/setup-python` : installation de Python 3.12
3. installation de MkDocs Material
4. ajout du marqueur `*` pour les branches non principales et détection automatique des nouveaux suivis de sprint (`docs/monitoring/sprint-*.md` ajoutés à la navigation sans modifier `mkdocs.yml`)
5. copie du PDF versionné (`documentation.pdf` à la racine) vers `docs/documentation.pdf` pour exposition sur le site
6. exécution de `mkdocs build`
7. publication d’un artefact de prévisualisation pour validation
8. publication sur GitHub Pages uniquement sur `main`

### Étapes du job `deploy`

Le job `deploy` dépend du job `build` et ne se lance que pour `main`. Il publie le contenu sur l’environnement GitHub Pages via :

- `actions/deploy-pages@v4`

Cela permet d’exposer la documentation officielle en ligne, selon l’URL définie par le dépôt GitHub Pages.

## Résultat attendu

Après un push ou un pull request sur `main` ou `dev`, GitHub Actions reconstruit la documentation. Les branches hors `main` affichent un suffixe `*` pour indiquer qu’il s’agit d’une version de développement. Seule la branche `main` est déployée publiquement sur GitHub Pages.

## Bonnes pratiques

- garder les fichiers Markdown dans `docs/` cohérents avec la structure du projet
- mettre à jour `mkdocs.yml` si un nouveau chapitre ou une nouvelle page est ajouté
- valider le build localement avec `mkdocs build` avant de pousser

## Commande de vérification locale

Pour construire le site en local :

```bash
cp documentation.pdf docs/documentation.pdf # expose le lien PDF de l'accueil
mkdocs build
```

Cette commande crée le dossier `site/`, qui est le même contenu que celui déployé sur GitHub Pages.
