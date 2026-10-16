# Sprint N°6 : Refonte ergonomique TUI (Gitelligence), fiabilisation des flux et robustesse LLM

User Stories :

- US 6.1 : En tant que développeur, je restructure le flux interactif de commit (`ShowCommitFlow`) du CLI Go en une architecture événementielle (`Event Loop`) découplée de la génération IA (`generateCommit`), et j'introduis la fonction partagée `ReadChoice` avec réécriture ANSI in-situ (`\033[1A\r` + `\033[K`) afin d'absorber les appuis sur Entrée à vide sans déclencher d'appel LLM parasite ni faire défiler l'invite du menu en cascade.

- US 6.2 : En tant que développeur, je conçois et standardise le composant d'affichage des erreurs `PrintErrorCard` dans `src/cli/ui/box.go` (encadré rouge stylisé, calcul de retour à la ligne automatique `wrapText`, message explicite et pistes de résolution) et je supprime définitivement l'ancienne fonction `PrintError` ainsi que le contournement technique de `commit_view.go` consistant à injecter un faux commit de type `"erreur"`.

- US 6.3 : En tant que développeur, je sécurise le parcours utilisateur en interdisant et en masquant contextuellement l'action de validation `[V]` en cas d'erreur de génération IA ou d'extraction Git dans l'écran de commit (restreignant le choix à `(R/M/A)`), et j'applique ce même garde-fous sur l'écran des release notes en masquant l'action de sauvegarde `[S]` en cas d'échec.

- US 6.4 : En tant que développeur, je restreins l'accès aux fonctionnalités de simulation sous le drapeau d'exécution `--dev` dans le CLI Go (`main.go`), masquant totalement les options Mock IA et Mock Interface de l'écran de configuration pour l'utilisateur final standard, tout en maintenant leur configuration opérationnelle via les variables d'environnement du `.env`.

- US 6.5 : En tant que développeur, j'épure la pile technique en supprimant définitivement le paramètre de langue (`APP_LANGUAGE` / champ `Language`) devenu inutile à travers l'interface CLI Go, le module bridge (`client.go`, `env.go`, `python_client.go`, `mock_client.go`), le point d'entrée Python `service/core.py`, le fichier de configuration `.env.exemple` et la suite de tests automatisés.

- US 6.6 : En tant que développeur, j'officialise le renommage de l'application en **Gitelligence** dans l'ensemble de la bannière TUI (`⚡ GITELLIGENCE ⚡ - Assistant Git Intelligent`), les messages utilisateur, la documentation, ainsi que la variable d'environnement de résolution racine `GITELLIGENCE_APP_ROOT` (avec repli rétrocompatible sur `SMART_COMMIT_APP_ROOT`).

- US 6.7 : En tant que développeur, je fiabilise le respect de la norme Conventional Commits par le LLM en implémentant un mécanisme déterministe d'override de type (`_extract_forced_type` et `_apply_forced_type` dans `ollama_client/llm.py`) qui force le type demandé par l'utilisateur lors d'un feedback explicite sans altérer le résumé ou le corps généré par l'IA, validé par 3 nouveaux scénarios golden diffs et portant la suite de tests Python à 52 tests unitaires passés.


- US 6.8 : En tant que développeur, j'accélère les pipelines d'intégration continue GitHub Actions (`.github/workflows/ci.yml`) en intégrant la mise en cache des dépendances pip (`cache: pip`), et j'assure la mise à jour continue du site MkDocs et du livrable `documentation.pdf`.

- US 6.9 : En tant que développeur, je finalise la conception et la livraison du poster de présentation pour le client, synthétisant l'architecture Go/Python, les métriques d'évaluation LLM et la proposition de valeur de Gitelligence.


Livrables / DoR & DoD :

- Interface TUI Gitelligence restructurée : architecture Event Loop sur `ShowCommitFlow`, utilitaire partagé `ReadChoice` avec réécriture ANSI in-situ, élimination des doubles exécutions LLM et des duplications d'invites sur appui d'Entrée vide.
- Composant générique `PrintErrorCard` déployé sur tous les écrans du CLI Go (`commit_view.go`, `release_notes_view.go`, `config_view.go`, `main.go`), code mort `PrintError` purgé et suppression définitive du hack de faux commit d'erreur.
- Verrouillage contextuel des actions en cas d'erreur : masquage de `[V]` sur les commits et de `[S]` sur les release notes empêchant toute validation ou sauvegarde erronée.
- Mode développeur `--dev` opérationnel : isolation et masquage des modes de simulation (Mock Interface et Mock IA) hors du flux utilisateur standard.
- Nettoyage architectural : suppression intégrale de la gestion multilingue obsolète (`APP_LANGUAGE`) du frontend Go jusqu'au backend Python.
- Marque Gitelligence déployée sur tout le projet : bannières, variables d'environnement (`GITELLIGENCE_APP_ROOT`), CLI et documentation MkDocs.
- Mécanisme d'override déterministe de type de commit via feedback intégré dans `ollama_client/llm.py`, validé à 100% sur 9 golden diffs complexes et couvert par 52 tests unitaires Python (hors slow).
- Pipeline CI/CD optimisé avec mise en cache pip, validation conjointe des tests Go et Python, et synchronisation du livrable `documentation.pdf`.
- Livrable poster de soutenance validé.
- Guide technique sur la gestion des environnements virtuels Python (`guide_environnements_python.md`) rédigé, et cadrage des prochaines itérations (notifications de fin d'inférence, refonte release notes).
