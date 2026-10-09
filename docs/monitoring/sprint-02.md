# Sprint N°3 : Intégration backend/CLI et fiabilisation du client LLM

User Stories :

- US 3.1 : En tant que développeur, j'écris le point d'entrée `service/core.py` qui relie le CLI Go au backend Python (statut, diff, génération de commit, application du commit, push, release notes, configuration).

- US 3.2 : En tant que développeur, j'unifie la lecture de la configuration (`.env` partagé) afin que tous les membres de l'équipe utilisent la même source de vérité pour l'URL Ollama, le modèle et les timeouts.

- US 3.3 : En tant que développeur, j'ajoute une vérification rapide de la joignabilité du serveur Ollama avant de lancer une génération, pour distinguer un problème réseau (hors réseau IUT) d'une vraie erreur du modèle.

- US 3.4 : En tant que développeur, je corrige la fenêtre de contexte utilisée par Ollama afin d'éviter une troncature silencieuse du diff sur les cas volumineux.

- US 3.5 : En tant que développeur, j'affine le prompt système du client LLM pour réduire les confusions de type Conventional Commits (`fix`/`feat`, `chore`/`build`) observées en mesure.

- US 3.6 : En tant que développeur, je relis le code d'intégration backend/CLI et transmets les anomalies trouvées avant fusion.

- US 3.7 : En tant que développeur, je complète la documentation miroir manquante (tests réseau, script de métriques) pour rester cohérent avec la convention de documentation de l'équipe.

- US 3.8 : En tant que développeur, j'identifie que le déploiement de la documentation utilisateur ne s'exécute jamais réellement, faute de déclencheur CI adapté à notre flux de travail (`Dev` plutôt que `main`).

Livrables / DoR & DoD :

- `service/core.py` fonctionnel, relié au module Git et au module LLM, mergé sur `Dev`.
- Configuration centralisée dans un fichier `.env` partagé, lu par tous les modules Python.
- Client Ollama fiabilisé : ping de disponibilité, fenêtre de contexte explicite, prompt corrigé, revalidé sur des scénarios réels (golden files + fixtures Git réelles).
- Documentation technique à jour pour les nouveaux fichiers de test et de mesure.
- Anomalie de déploiement de la documentation utilisateur identifiée et documentée, correction reportée au sprint suivant.