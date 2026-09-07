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
### Auteurs du sujet :
- **Rémi Cozot**
- **Rémi Synave**

## Langage utilisés :
- **GO en front-end**
- **Python en back-end**


### Auteurs : 
- **Dorian HEMBERT**
- **Maxence LAURENCE**
- **Enzo CIUFFA**