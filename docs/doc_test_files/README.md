# Présentation des tests

Cette section décrit la documentation technique associée aux tests du projet.

## Objet

Le dossier `test/` contient les tests unitaires, d’intégration et de situation. Cette documentation permet de comprendre le rôle de chaque test et la stratégie de validation du projet.

## Types de tests présents

- tests unitaires : vérification des fonctions et composants isolés
- tests de wrapper Git : validation de l’interaction avec Git
- tests Ollama LLM : vérification des appels et réponses du modèle
- golden diffs : tests sur des sorties attendues
- tests de situation : scénarios réalistes du projet

## Organisation

La structure des fichiers de documentation suit celle du dossier `test/` pour faciliter la lecture et la maintenance.

## Rôle de cette documentation

Chaque test ou groupe de tests possède une page dédiée afin de :

- expliquer le but du test
- détailler le scénario couvert
- documenter les résultats attendus
- faciliter le suivi qualité du projet

## À retenir

Les tests sont une partie essentielle de la fiabilité du projet. Leur documentation doit rester à jour pour refléter les cas de validation réellement mis en place dans le code.
