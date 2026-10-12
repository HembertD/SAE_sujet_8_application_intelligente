# Sprint N°5 : Diagnostic réseau unifié, stabilisation CI/CD, version stable sur main et améliorations UX

User Stories :

- US 5.1 : En tant que développeur, j'unifie la vérification de connectivité avec le serveur Ollama via une fonction centralisée `check_ollama_reachable` (alias `ping_ollama`) et une structure enrichie `PingResult` dans `ollama_client/llm.py`, fournissant la latence réseau en millisecondes, le statut d'accessibilité, la liste des modèles installés et une compatibilité de déballage en tuple.

- US 5.2 : En tant que développeur, je sécurise le point d'entrée `service/core.py` en intégrant ce ping unifié avant toute requête et en appliquant un mécanisme de fail-fast strict sur `action_release_notes` afin de renvoyer une erreur explicite au CLI Go plutôt que d'utiliser silencieusement un modèle déterministe lors d'une rupture de connexion.

- US 5.3 : En tant que développeur, j'expose explicitement la variable `OLLAMA_NUM_CTX=32768` dans le fichier `.env.exemple` partagé pour aligner les environnements sur la fenêtre de contexte maximale du modèle et prévenir la troncature silencieuse des diffs volumineux.

- US 5.4 : En tant que développeur, j'écris une suite de tests unitaires pour le diagnostic Ollama (`test/test_ollama_llm.py` et `test/test_core.py`), couvrant la mesure de latence, l'extraction des modèles, la compatibilité d'interface et l'interruption précoce sur indisponibilité réseau, portant la suite automatisée à 45 tests unitaires passés.

- US 5.5 : En tant que développeur, je stabilise et sécurise le pipeline CI GitHub Actions (`.github/workflows/ci.yml`) en affinant les déclencheurs de pull request et de push vers `dev` et `main`, en restreignant les permissions par défaut, et en épinglant les dépendances MkDocs (`mkdocs==1.6.0`, `mkdocs-material==9.5.34`, `PyYAML==6.0.2`) pour garantir des builds déterministes.

- US 5.6 : En tant que développeur, je résous les conflits de nommage sur le backend en supprimant `src/logging.py` qui masquait la bibliothèque standard Python, j'assainis le dépôt Git (`.gitignore`, suppression de fichiers vides) et j'automatise l'activation et la publication de GitHub Pages via l'API d'enablement GitHub.

- US 5.7 : En tant que développeur, je déploie la première version stable officielle de l'application sur la branche `main` (PR #36 et PR #37) après validation conjointe de tous les contrôles qualité CI (tests Go, tests Python mockés, génération du PDF et site MkDocs).

- US 5.8 : En tant que développeur, je conçois un composant standardisé d'affichage des cartes d'erreur (`PrintErrorCard`) dans `src/cli/ui/` afin de supprimer le contournement technique de `commit_view.go` (injection d'un faux commit de type `"erreur"`) et harmoniser la restitution des erreurs dans l'ensemble des vues de l'interface Go.

- US 5.9 : En tant que développeur, je renforce les contrôles de flux de travail en verrouillant les choix de validation du CLI Go en cas d'erreur de génération et en préparant une boucle de nouvel essai (retry loop) pour rattraper les non-conformités de formatage de l'IA.

- US 5.10 : En tant que développeur, j'améliore l'expérience utilisateur (UX) du CLI Go en intégrant des indicateurs ou notifications de fin de traitement lors des requêtes d'analyse du modèle afin d'avertir clairement l'utilisateur dès que la génération est terminée.

- US 5.11 : En tant que développeur, je prépare la réalisation du poster demandé par le client afin de synthétiser les fonctionnalités, l'architecture technique et les bénéfices de l'application en vue de la soutenance.

Livrables / DoR & DoD :

- Système de diagnostic et de ping réseau Ollama unifié (`PingResult`, `check_ollama_reachable`, `ping_ollama`) opérationnel, testé unitairement (45 tests unitaires Python validés hors-ligne) et documenté dans les fiches miroirs (`llm.md`, `core.md`, `test_core.md`, `test_ollama_llm.md`).
- Comportement fail-fast strict déployé sur la génération des release notes dans `service/core.py` en cas d'indisponibilité du serveur IA.
- Variable d'environnement `OLLAMA_NUM_CTX` officialisée et documentée dans `.env.exemple`.
- Pipeline CI/CD GitHub Actions fiabilisé, allégé et reproductible : exécution parallèle des tests Go et Python déconnectés, permissions restreintes, épinglage des dépendances MkDocs et déploiement continu GitHub Pages sans blocage.
- Dépôt nettoyé et assaini : suppression du conflit stdlib `src/logging.py`, nettoyage des fichiers vides et réconciliation du `.gitignore`.
- Première version stable officielle validée et livrée sur la branche `main` (PR #36 & PR #37).
- Spécification et architecture du composant d'erreur TUI unifié (`PrintErrorCard`) et des garde-fous de validation définies pour l'interface CLI Go.
- Préparation et maquettage du poster de présentation demandé par le client pour la soutenance finale.
- Suivi de projet synchronisé sur la branche `sprints` et aligné avec les exigences du jalon du 16 octobre.
