# Sprint N°8 : Sélecteur interactif de commits (Release Notes), fiabilisation Docker Ollama (gemma4:12b) et pilotage TUI de l'environnement Python

User Stories :

- US 8.1 : En tant que développeur, je conçois et intègre un composant interactif de sélection en mode brut TTY (`SelectCommit` dans `src/cli/ui/selector.go`) avec capture dynamique des touches clavier (`stty cbreak -echo`), navigation fluide (touches fléchées, raccourcis vim `j`/`k`, touches Début/Fin), et pagination par fenêtre déroulante avec indicateurs visuels `▲` / `▼` afin d'éliminer toute saisie textuelle manuelle sujette aux erreurs de syntaxe Git.

- US 8.2 : En tant que développeur, je refonds l'interface utilisateur de génération des release notes (`src/cli/views/release_notes_view.go`) pour articuler le flux autour d'une double sélection interactive au clavier dans l'historique complet du dépôt (point de départ `From` et point d'arrivée `To`), affichant pour chaque révision son hash court, sa date relative et son intitulé, avec réécriture propre in-situ.

- US 8.3 : En tant que développeur, j'expose l'historique récent des commits du dépôt à travers le backend Python (`GitWrapper.get_recent_commits`, nouvelle action `commits` dans `service/core.py`) et le bridge Go (`src/cli/bridge/client.go`, `python_client.go`, `mock_client.go`, `models/types.go`) via les contrats de données partagés `CommitItem` et `CommitsResponse`, en garantissant un support hors-ligne réaliste pour les tests et la démonstration.

- US 8.4 : En tant que développeur, je fiabilise la méthode `get_commits_between_tags` dans `GitWrapper` en assurant un tri chronologique strict et une gestion rigoureuse de l'inclusivité des bornes de révision, empêchant l'inversion ou la troncature de la liste des commits transmise au modèle de langage.

- US 8.5 : En tant que développeur, je calibre le modèle LLM local par défaut sur `gemma4:12b` (~8 Go) à travers l'ensemble de la pile technique (`.env.exemple`, `config.py`, `llm.py`, bridge Go, mocks et tests).

- US 8.6 : En tant que développeur, je durcis le conteneur Docker Ollama local (`Dockerfile`, `docker-compose.yml`, `entrypoint.sh`) en remplaçant la sonde d'attente et de healthcheck `curl` (absente de l'image officielle) par la commande native `ollama list`, en propageant dynamiquement le modèle configuré via `OLLAMA_MODEL`, et en intégrant un trap sur les signaux `SIGTERM`/`SIGINT` pour assurer un arrêt propre et sans corruption du serveur d'inférence.

- US 8.7 : En tant que développeur, je sécurise l'intégration Docker dans le CLI Go (`client.go`, `env.go`, `python_client.go`, `config_view.go`) en déléguant le démarrage/arrêt du conteneur au client actif (garantissant que le `MockClient` n'exécute aucune commande Docker réelle), en centralisant l'URL locale effective via `EffectiveOllamaBaseURL()` (`http://localhost:11434`), et en préservant le drapeau `USE_DOCKER_LLM` lors de la réécriture du `.env`.

- US 8.8 : En tant que développeur, je raccorde l'environnement Python autonome dans l'écran de configuration du CLI Go (Carte Trello #82 / #54, `src/cli/views/config_view.go`, `src/cli/bridge/env.go`) en ajoutant l'action interactive `[P]` pour tester la résolution et afficher la version de l'interpréteur (`python --version`), ainsi que la configuration dynamique et la persistance de `PYTHON_BIN` via l'action de modification `[M]` (avec vérification d'accessibilité du binaire et réinitialisation automatique en mode `.venv` via `auto`).

- US 8.9 : En tant que développeur, j'étends la suite de tests automatisés Go (`test/cli/bridge_test.go`) avec 112 nouvelles assertions couvrant la résolution d'URL effective Docker, la priorisation des variables d'environnement système sur le fichier `.env`, l'isolation stricte du mode mock face aux commandes Docker et la persistance des paramètres, tout en couvrant l'action `commits` dans `test/test_core.py`.

- US 8.10 : En tant que développeur, j'enrichis la documentation technique MkDocs (`mkdocs.yml`) avec une section Docker dédiée, je rédige les fiches miroirs des nouveaux composants (`selector.md`, `Dockerfile.md`, `docker-compose.md`, `entrypoint.md`), j'aligne le schéma d'architecture global (`doc_architecture.md`), le guide de démarrage rapide (`README.md`) et j'actualise le livrable PDF `documentation.pdf`.

- US 8.11 : En tant que développeur, je formalise sur le tableau Trello les User Stories et critères d'acceptation détaillés des prochains jalons : drapeau d'aide universel `--help` (carte #90), exécution asynchrone non-bloquante de l'IA avec interruption réactive (carte #92), garde-fous de non-destructivité des modifications de fichiers avec confirmation interactive obligatoire (carte #91), et téléchargement automatique de modèles Ollama (carte #89).


Livrables / DoR & DoD :

- Composant interactif `SelectCommit` opérationnel en mode brut TTY (`src/cli/ui/selector.go`), avec capture clavier réactive, pagination déroulante avec indicateurs `▲` / `▼` et réinitialisation propre du terminal.
- Écran des release notes (`release_notes_view.go`) entièrement refondu avec sélection interactive du commit initial et final, éliminant toute saisie textuelle de révisions Git.
- Endpoint backend `commits` (`service/core.py`) et méthode `get_recent_commits` (`git_wrapper.py`) intégrés, supportés par le `MockClient` pour les tests hors-ligne et validés par `test_core.py`.
- Méthode `get_commits_between_tags` durcie sur l'inclusivité et le tri chronologique.
- Modèle local Ollama réajusté sur `gemma4:12b` (~8 Go) dans l'ensemble de la stack (fichiers de configuration, Go, Python, mocks, tests et documentation), écartant le modèle 26B trop lourd.
- Stack de conteneurisation Docker Ollama fiabilisée : healthcheck et boucle d'attente basés sur `ollama list` (suppression de la dépendance à `curl`), trap `SIGTERM`/`SIGINT` pour arrêt propre, et injection dynamique de `OLLAMA_MODEL`.
- Mode Docker dans le CLI Go découplé et sécurisé : gestion via l'interface `BackendClient`, fallback propre en mode mock, résolution centralisée via `EffectiveOllamaBaseURL()`, et préservation de `USE_DOCKER_LLM` dans le `.env`.
- Raccordement de l'environnement Python dans l'interface CLI Go (carte Trello #82 / #54) : action `[P]` pour vérification en temps réel de l'interpréteur et de sa version, configuration interactive de `PYTHON_BIN` avec persistance `.env` via `WritePythonBinSetting`, et option `auto`.
- Suite de tests Go enrichie dans `test/cli/bridge_test.go` (112 nouvelles assertions) et tests Python dans `test_core.py`, maintenant 100% de tests au vert.
- Documentation MkDocs enrichie (nouvelle section Docker dans `mkdocs.yml`, fiches miroirs `selector.md`, `Dockerfile.md`, `docker-compose.md`, `entrypoint.md`, `doc_architecture.md`, `README.md` et livrable `documentation.pdf` actualisé).
- Spécification et priorisation agile des cartes Trello du backlog (#89, #90, #91, #92, #94) pour le sprint suivant.
