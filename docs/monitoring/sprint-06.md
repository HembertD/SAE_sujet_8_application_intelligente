# Sprint N°7 : Notifications desktop, environnement autonome, conteneurisation Docker et résolution de conflits LLM

User Stories :

- US 7.1 : En tant que développeur, j'implémente un système de notifications de bureau multi-plateforme (Linux via D-Bus natif, macOS via Notification Center, Windows via Toast) déclenché de manière asynchrone dès la fin des inférences IA pour les messages de commit et les release notes, entièrement paramétrable et débrayable via la variable `ENABLE_NOTIFICATIONS` dans le `.env` et la vue de configuration du CLI Go.

- US 7.2 : En tant que développeur, je conçois un script d'initialisation d'environnement virtuel autonome (`scripts/setup_python_env.py`) installant automatiquement les dépendances requises dans un dossier `.venv/` à la racine de l'application, et j'intègre dans le client Go (`src/cli/bridge/python_client.go`) une cascade de résolution multi-plateforme avec garde fail-fast strict au démarrage guidant l'utilisateur en cas d'interpréteur manquant.

- US 7.3 : En tant que développeur, je raccorde l'action interactive d'imposition de type Conventional Commit (`[T]`) dans le menu de décision du CLI Go (`src/cli/views/commit_view.go`), permettant à l'utilisateur de sélectionner explicitement l'un des 9 types autorisés avec une consigne optionnelle, répercutée de façon déterministe sur le modèle et alignée sur le `MockClient`.

- US 7.4 : En tant que développeur, je développe un module de proposition de résolution des conflits de merge assistée par l'IA (`ollama_client/conflict_resolver.py`), articulé autour des contrats de données partagés `ConflictSection` et `ConflictResolution`, d'un prompt système dédié rejetant tout marqueur résiduel, d'une boucle de validation avec retry et d'une prise en charge du mode mock hors-ligne.

- US 7.5 : En tant que développeur, j'unifie le socle d'exécution HTTP LLM (`call_chat_json` dans `ollama_client/llm.py`) et je refonds le moteur de génération des release notes (`ollama_client/release_notes.py`) pour éliminer l'appel HTTP brut dans `service/core.py`, garantissant un schéma JSON contraint (`{"markdown": str}`), le respect de la fenêtre de contexte `OLLAMA_NUM_CTX` et un prompt système externalisé.

- US 7.6 : En tant que développeur, je sécurise la génération des release notes en supprimant définitivement le repli silencieux vers un faux payload déterministe lors des échecs d'inférence en mode réel, renvoyant une exception explicite `ReleaseNotesError` au CLI Go afin de garantir la transparence pour l'utilisateur.

- US 7.7 : En tant que développeur, j'augmente la résilience du client LLM face aux aléas de génération en étendant le mécanisme de nouvel essai unique (retry) aux réponses brutes non conformes au format JSON via `MalformedResponseError`.

- US 7.8 : En tant que développeur, je fiabilise l'analyse Git dans `GitWrapper` en corrigeant la fausse détection de fichiers binaires sur le texte contenant `"Binary files"`, et en distinguant rigoureusement les fichiers indexés vides de l'absence de modifications en staging.

- US 7.9 : En tant que développeur, j'intègre une solution de conteneurisation locale du serveur Ollama avec le modèle `gemma4:26b` via Docker (`Dockerfile`, `docker-compose.yml`, `entrypoint.sh`), pilotable depuis l'interface CLI Go grâce au support de la variable `USE_DOCKER_LLM`.

- US 7.10 : En tant que développeur, j'assainis et consolide la documentation technique miroir (`docs/doc_src_files/`, `docs/doc_test_files/`, `doc_architecture.md`, `README.md`) pour refléter les nouveaux modules (notifications, résolveur de conflits, environnement autonome, Docker), et je procède au nettoyage des branches obsolètes sur le dépôt distant.

Livrables / DoR & DoD :

- Sous-système de notifications de bureau multi-plateforme opérationnel (`src/cli/ui/notification.go`), testé unitairement (`test/cli/notification_test.go`), non-bloquant et intégré aux écrans de commit et de release notes.
- Système d'environnement autonome complet : script `setup_python_env.py`, détection automatique de `<appRoot>/.venv`, variable `PYTHON_BIN`, support multi-plateforme des séparateurs de chemins (`os.PathListSeparator`) et arrêt sécurisé au démarrage du CLI si l'environnement est absent.
- Action interactive `[T] Type imposé` intégrée au CLI Go, validée par des tests unitaires (`bridge_test.go`) et alignée sur le mode mock.
- Moteur de résolution de conflits de merge (`conflict_resolver.py`, `models.py`) opérationnel, validé par 9 tests unitaires hors-ligne.
- Génération des release notes harmonisée sur `call_chat_json` avec prompt externalisé (`release_notes_system.txt`), sortie JSON stricte et suppression totale du fallback silencieux.
- Fiabilisation du parsing LLM avec retry sur réponse non-JSON (`MalformedResponseError`).
- Détection Git durcie (`GitWrapper`) sur les fichiers binaires et staged vides.
- Stack Docker locale (`docker-compose.yml`, `Dockerfile`, `entrypoint.sh`) prête à l'emploi et connectable via `USE_DOCKER_LLM`.
- Suite automatisée validée : 72 tests unitaires Python (hors slow) et 100% des tests Go réussis.
- Documentation technique et navigation MkDocs intégralement synchronisées avec les évolutions du sprint.
