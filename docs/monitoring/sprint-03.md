# Sprint N°4 : Simulation hors-ligne (Mock IA), robustesse backend et déploiement continu de la documentation

User Stories :

- US 4.1 : En tant que développeur, je développe un module de mock IA (`mock_llm.py`) fournissant des réponses déterministes et statiques (génération de message de commit, release notes, ping de disponibilité) afin de permettre le développement et l'exécution hors-ligne sans dépendre du serveur Ollama de l'IUT.

- US 4.2 : En tant que développeur, j'adapte l'interface CLI Go en remplaçant l'option obsolète `--demo` par deux drapeaux distincts (`--mockInterface` et `--mockIA`), et j'ajoute leur gestion interactive dans le menu de configuration ainsi que la synchronisation de la variable `MOCK_AI` dans le fichier `.env`.

- US 4.3 : En tant que développeur, je sécurise le point d'entrée `service/core.py` en évitant l'arrêt brutal par `sys.exit(1)` sur action inconnue au profit d'un retour JSON exploitable par le CLI Go, et j'exploite la méthode publique `get_staged_diff` du wrapper Git.

- US 4.4 : En tant que développeur, j'écris une suite de tests unitaires complète pour le mock IA (`test_mock_ai.py`) isolée par `conftest.py`, ainsi qu'un test unitaire Go pour la persistance conjointe des options de mock (`TestSaveConfigBothMocks`).

- US 4.5 : En tant que développeur, je résous l'anomalie de déploiement de la documentation identifiée au sprint précédent en automatisant la publication du site MkDocs sur GitHub Pages lors des pushs et pull requests vers `Dev` et `main`.

- US 4.6 : En tant que développeur, je mets en place la génération automatisée d'un livrable PDF de documentation (`documentation.pdf`) via un script dédié (`generate_documentation_pdf.py`), un hook git de pré-commit et son intégration dans le workflow CI.

- US 4.7 : En tant que développeur, j'intègre l'exécution automatique des tests Go dans le pipeline CI (`.github/workflows/ci.yml`) à chaque push et pull request.

- US 4.8 : En tant que développeur, j'actualise l'ensemble de la documentation technique miroir (`mock_llm.md`, `test_mock_ai.md`, `llm.md`, `deploiement_documentation.md`) et les guides utilisateurs pour refléter les évolutions architecturales et fonctionnelles du projet.

- US 4.9 : En tant que développeur, je restaure le job de tests unitaires Python (`python-tests`) dans le workflow CI (`.github/workflows/ci.yml`), en ajoutant la dépendance `pytest>=7.0.0` dans `requirements.txt` et en actualisant `docs/deploiement_tests_workflows.md` pour exécuter automatiquement les 40 tests unitaires mockés sans nécessiter de connexion au réseau IUT.

Livrables / DoR & DoD :

- Module `MockLLMClient` (`src/git_commit_release_notes_generator/ollama_client/mock_llm.py`) fonctionnel et intégré, activable via la variable `MOCK_AI` du `.env`.
- CLI Go mis à niveau avec prise en charge des flags `--mockInterface` et `--mockIA`, affichage des statuts dans le menu principal et gestion dans le menu de configuration.
- Point d'entrée `service/core.py` fiabilisé : retours d'erreurs en JSON sans arrêt non contrôlé du sous-processus et harmonisation des appels au wrapper Git (`get_staged_diff`).
- Job CI `python-tests` restauré dans `.github/workflows/ci.yml` (avec `pytest>=7.0.0` dans `requirements.txt` et mise à jour de `docs/deploiement_tests_workflows.md`), exécutant les 40 tests unitaires Python hors-ligne en parallèle des tests Go (`bridge_test.go`).
- Déploiement automatique du site de documentation MkDocs sur GitHub Pages opérationnel sur les branches `Dev` et `main` (anomalie du Sprint N°3 résolue).
- Génération automatisée du livrable `documentation.pdf` intégrée à la CI et outillée localement avec un hook Git pre-commit.
- Documentation technique miroir et guides utilisateurs synchronisés avec l'arborescence du projet.
- Pistes d'amélioration UX (verrouillage de validation en cas d'erreur de génération, notifications de fin de traitement de l'IA) identifiées sur le Trello et planifiées pour le sprint suivant.
