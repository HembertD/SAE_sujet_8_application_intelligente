# Vue Configuration

Fichier source associé : src/cli/views/config_view.go

## Description

Ce module permet de visualiser et de modifier les paramètres opérationnels de l'application (URL Ollama, modèle LLM, timeout, notifications de bureau, option LLM via Docker, interpréteur Python) et de diagnostiquer la connectivité réseau avec le serveur de l'IUT ainsi que l'état de l'environnement Python.

## Responsabilités

- afficher les paramètres actuellement appliqués issus du fichier d'environnement `.env` situé à la racine du projet (URL, Modèle, Timeout, Notifications IA, statuts distincts `Mock Interface`, `Mock IA`, `LLM via Docker` et `Conteneur Docker`, ainsi que l'interpréteur Python résolu avec indication du mode `.venv` ou `PYTHON_BIN`) ; en mode Docker, l'URL affichée est l'adresse réellement utilisée (`http://localhost:11434`) ;
- afficher l'état réel du conteneur Ollama (`En cours`, `Arrêté` ou `Indisponible`) en interrogeant `docker compose ps -q ollama`, indépendamment du réglage `USE_DOCKER_LLM` ;
- **Action [T] (Test Ping)** : exécuter un diagnostic réseau vers l'API Ollama réellement utilisée (`http://localhost:11434` en mode Docker, sinon l'URL du `.env`, ou simulation instantanée si Mock IA actif), mesurer la latence en millisecondes et lister l'ensemble des modèles d'IA installés sur le serveur ;
- **Action [P] (Vérification Python)** : tester en profondeur l'environnement Python, valider la présence de toutes les dépendances requises (GitPython, PyYAML), confirmer le chargement effectif du backend applicatif et afficher la version exacte (`python --version`) ainsi que le mode d'attribution (`.venv` ou `PYTHON_BIN`) ;
- **Action [M] (Modification)** : permettre la saisie de nouvelles valeurs et persister les modifications dans le fichier `.env` à la racine de l'application sans aucune recompilation du binaire (y compris les options `Notifications IA`, `Mock Interface`, `Mock IA`, `LLM via Docker` et `PYTHON_BIN` avec vérification préalable systématique du candidat Python : si la sonde échoue ou que des dépendances manquent, la modification est refusée et l'ancienne valeur est conservée sans modification du `.env`) ;
- après sauvegarde, si `LLM via Docker` est activé : proposer de lancer le conteneur Ollama via `docker compose up -d --build` avec une barre de progression, puis rappeler que le premier téléchargement du modèle peut prendre plusieurs minutes (`docker compose logs -f ollama`) ;
- après sauvegarde, si `LLM via Docker` est désactivé : proposer d'arrêter le conteneur Ollama via `docker compose down` ;
- **Action [D] (Docker LLM)** : basculer directement le mode Docker sans passer par `[M]` : si le mode est désactivé, proposer de l'activer et de démarrer le conteneur (`docker compose up -d --build`) ; s'il est activé, proposer de le désactiver et d'arrêter le conteneur (`docker compose down`). L'état réel du conteneur est vérifié avant l'action pour éviter les commandes redondantes ; la configuration `.env` n'est mise à jour qu'après le succès de l'opération Docker (ou avec un avertissement explicite si l'état Docker est indisponible lors d'une désactivation) ;
- **Action [R]** : retour au menu principal.

## Points clés

- **Conformité stricte aux exigences d'évaluation** : respect de la règle d'or "tout doit être dans un `.env`, aucune recompilation requise lors d'un changement de modèle ou d'adresse" ;
- **Diagnostic intégré complet** : permet de valider instantanément si le serveur de l'IUT (`10.22.28.190:11434`) répond et si l'environnement d'exécution Python autonome ou personnalisé est opérationnel ;
- **Pilotage Docker sans détour** : la ligne `Conteneur Docker` et l'action `[D]` rendent l'état et le cycle de vie du conteneur Ollama visibles et actionnables directement depuis l'écran Configuration, sans passer par l'édition du `.env` avec `[M]` ;
- **Saisie sécurisée** : utilisation de `ReadChoice` pour éviter toute sortie accidentelle lors d'un appui sur Entrée à vide (retour exigeant explicitement la touche `[R]`).

## Rôle dans l’architecture

Interface d'administration et de diagnostic du front-end, garantissant la configurabilité à chaud du système.

## Commentaires de fonction (dupliqués du code source)

- `ShowConfigFlow()` : Affiche et gère le flux interactif de configuration, incluant l'affichage de l'état du conteneur, l'activation/désactivation de `USE_DOCKER_LLM` et le lancement/arrêt du conteneur Ollama Docker avec barre de progression.
- `toggleDockerLLM()` : Active ou désactive le mode Docker et synchronise le conteneur Ollama via `StartDockerOllama` / `StopDockerOllama`, puis persiste `USE_DOCKER_LLM` dans le `.env`.
- `printDockerUnavailable()` : Affiche un diagnostic lorsque l'état du conteneur ne peut pas être interrogé (Docker arrêté, binaire absent, `docker-compose.yml` manquant).
- `dockerErrorDetail()` : Extrait le message d'erreur le plus pertinent d'une action Docker.
