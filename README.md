# SAE - Application intelligente

## Présentation

```
L'objectif de cette SAÉ est de développer un outil en ligne de commande (CLI) ou une interface graphique locale capable d'automatiser la rédaction des messages de commit et la génération de notes de version (release notes).
Les développeurs négligent souvent la qualité des messages de commit, ce qui rend l'historique du projet difficile à lire et la génération de changelogs fastidieuse. Le logiciel analysera les différences de code (git diff) pour déléguer cette tâche sémantique à un LLM.
```

## Fonctionnalités principales :
```
Analyse de code : exécution et parsing de la commande git diff pour extraire les modifications locales avant le commit.
Génération structurée : utilisation de gemma4:12b avec un prompt strict pour générer un message respectant la norme Conventional Commits (ex. feat: ajout du bouton de connexion).
Synthèse de version : extraction de l'historique entre deux tags Git et génération d'un changelog organisé par catégories (Corrections, Nouveautés, Régressions) au format Markdown.
Sécurité : filtrage pour exclure les fichiers binaires ou les données sensibles (clés d'API) avant l'envoi au modèle.
```
## Auteurs du sujet

- Rémi Cozot
- Rémi Synave

## Membres du projet

- Dorian HEMBERT
- Enzo CIUFFA
- Maxence LAURENCE

## Technologies

- Go pour le frontend / CLI
- Python pour le backend et la logique métier
- Git pour la gestion des changements
- MkDocs pour la documentation web
- Pandoc pour la génération du PDF de documentation
- GitHub Actions pour la CI/CD

## Documentation du dépôt

- [Documentation PDF complète](documentation.pdf)

Le projet contient plusieurs niveaux de documentation, répartis selon leur objectif.

### 1. Documentation générale et architecture

- [doc_architecture.md](doc_architecture.md) : architecture documentaire du projet et règle de correspondance entre le code et la documentation
- [docs/index.md](docs/index.md) : page d’accueil de la documentation MkDocs

### 2. Documentation technique du code source

La documentation technique du code est dans le dossier :

- [docs/doc_src_files](docs/doc_src_files)

Elle suit la même arborescence que le dossier source :

- [src](src)
- [docs/doc_src_files](docs/doc_src_files)

Exemples :

- [src/git_commit_release_notes_generator/config.py](src/git_commit_release_notes_generator/config.py)
- [docs/doc_src_files/git_commit_release_notes_generator/config.md](docs/doc_src_files/git_commit_release_notes_generator/config.md)

### 3. Documentation des tests

La documentation des tests et cas de validation est dans :

- [docs/doc_test_files](docs/doc_test_files)

Elle correspond aux fichiers du dossier tests :

- [test](test)
- [docs/doc_test_files](docs/doc_test_files)

### 4. Suivi de projet et sprints

La documentation de suivi est dans :

- [docs/monitoring](docs/monitoring)

Elle contient les comptes-rendus de sprint et le suivi d’avancement du projet.

### 5. CI, déploiement et documentation opérationnelle

- [docs/deploiement_documentation.md](docs/deploiement_documentation.md) : documentation du déploiement de la doc
- [docs/deploiement_tests_workflows.md](docs/deploiement_tests_workflows.md) : documentation de la CI / tests GitHub Actions
- [docs/explication_go_cli.md](docs/explication_go_cli.md) : explication du CLI Go

### 6. Documentation générée en PDF

Le PDF de documentation est produit automatiquement via GitHub Actions à partir de la structure MkDocs. Il est généré sur les branches principales lors des push et des pull requests.

La sortie attendue est un document unique de documentation, avec table des matières, dans un ordre cohérent selon la navigation MkDocs.

## Arborescence principale

```text
.
├── README.md
├── contributing.md
├── doc_architecture.md
├── mkdocs.yml
├── pytest.ini
├── requirements.txt
├── .github/
│   └── workflows/
│       └── ci.yml
├── src/
│   └── cli/
│   └── git_commit_release_notes_generator/
├── test/
├── docs/
│   ├── index.md
│   ├── deploiement_documentation.md
│   ├── deploiement_tests_workflows.md
│   ├── explication_go_cli.md
│   ├── monitoring/
│   ├── doc_src_files/
│   └── doc_test_files/
├── scripts/
│   └── run_metrics.py
└── .gitignore
```

voir plus dans le fichier [doc_architecture.md](doc_architecture.md)

## Règle de documentation

La documentation suit cette convention :

- le code applicatif est dans [src](src)
- la documentation technique correspondante est dans [docs/doc_src_files](docs/doc_src_files)
- les tests sont dans [test](test)
- leur documentation correspondante est dans [docs/doc_test_files](docs/doc_test_files)
- le suivi projet est dans [docs/monitoring](docs/monitoring)
- la documentation de déploiement et CI est dans [docs](docs)

## Liens utiles

- [Documentation MkDocs](docs/index.md)
- [Architecture de documentation](doc_architecture.md)
- [Déploiement de la documentation](docs/deploiement_documentation.md)
- [Tests automatisés](docs/deploiement_tests_workflows.md)
- [Projet GitHub](https://github.com/hembertd/SAE_sujet_8_application_intelligente)

## Trello

- https://trello.com/invite/b/6a9e6bc6e505a585bbe40bf8/ATTIa7cd65089f34fb9e4d5b7e8d388a9f602737B962/saesujet8applicationintelligente

## Note

La documentation est mise à jour au fil du projet. Si vous ajoutez un module ou un test, pensez à :

- ajouter la documentation correspondante dans le dossier adéquat
- mettre à jour la navigation MkDocs si nécessaire
- vérifier que le PDF de documentation se régénère correctement via GitHub Actions