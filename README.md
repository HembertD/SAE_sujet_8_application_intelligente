# Projet SAE application intélligente, sujet 8 

## Descriptif du projet :

```
L'objectif de cette SAÉ est de développer un outil en ligne de commande (CLI) ou une interface graphique locale capable d'automatiser la rédaction des messages de commit et la génération de notes de version (release notes).
Les développeurs négligent souvent la qualité des messages de commit, ce qui rend l'historique du projet difficile à lire et la génération de changelogs fastidieuse. Le logiciel analysera les différences de code (git diff) pour déléguer cette tâche sémantique à un LLM.

Fonctionnalités principales :
Analyse de code : exécution et parsing de la commande git diff pour extraire les modifications locales avant le commit.
Génération structurée : utilisation de gemma4:12b avec un prompt strict pour générer un message respectant la norme Conventional Commits (ex. feat: ajout du bouton de connexion).
Synthèse de version : extraction de l'historique entre deux tags Git et génération d'un changelog organisé par catégories (Corrections, Nouveautés, Régressions) au format Markdown.
Sécurité : filtrage pour exclure les fichiers binaires ou les données sensibles (clés d'API) avant l'envoi au modèle.
```
#### Auteurs du sujet :
- ***Rémi Cozot***
- ***Rémi Synave***

## Lien du Trello :
*https://trello.com/invite/b/6a9e6bc6e505a585bbe40bf8/ATTIa7cd65089f34fb9e4d5b7e8d388a9f602737B962/saesujet8applicationintelligente*

## Langages utilisés :
- **GO en front-end**
- **Python en back-end**

## Répartition des tâches

- **Dorian HEMBERT**

  * Mise en place et gestion du dépôt Git
  * Mise en place et configuration du système de Git Hooks
  * Détection et récupération des modifications avec `git diff`
  * Analyse et structuration des informations issues des modifications
  * Filtrage des fichiers binaires et des données sensibles

- **Enzo CIUFFA**

  * Conception et mise en place des prompts destinés au LLM
  * Encadrement des réponses de l'IA
  * Génération des messages de commit selon la norme Conventional Commits
  * Traitement et validation des réponses générées par le LLM

- **Maxence LAURENCE**

  * Développement du CLI
  * Intégration des différentes parties de l'application
  * Génération et affichage des notes de version (Release Notes)
  * Gestion du Trello
  * Tests du langage "Go"

### Auteurs : 
- **Dorian HEMBERT**
- **Maxence LAURENCE**
- **Enzo CIUFFA**