# Tests du client Ollama

Fichier source associé : test/test_ollama_llm.py

## Description

Ce module couvre la logique de validation et d’interaction avec le client Ollama. Il vérifie notamment le comportement du code qui transforme un diff en un message de commit conforme à la norme Conventional Commits.

## Ce qui est testé

- la génération d’un message de commit valide dès le premier essai ;
- le mécanisme de retry lorsqu’une réponse est invalide ;
- l’échec après plusieurs réponses invalides ;
- les règles de validation du message (`type`, `scope`, `subject`) ;
- la compatibilité avec les anciennes structures de données binaires.

## Stratégie de test

Les appels réseau réels sont évités par remplacement de `urllib.request.urlopen` avec un faux objet de réponse. Cela permet de simuler le comportement de Ollama sans dépendre d’un service distant.

## Cas importants

### génération réussie au premier essai

Le test vérifie qu’un payload bien formé est accepté et transformé en objet métier avec les bons champs : type, scope et subject.

### retry puis succès

Lorsque la première réponse est invalide, le système réessaie avec une seconde réponse. Cela confirme que la logique de récupération est bien mise en place.

### validation stricte du message

Le test contrôle que les sujets terminés par un point, ainsi que les types inconnus, sont rejetés pour rester conforme à la convention de messages attendue.

## Points de vigilance

- La validation doit rester stricte pour éviter des messages de commit peu lisibles ou incohérents.
- Le mécanisme de retry ne doit pas masquer des erreurs de fond ou des réponses non exploitables.
- La compatibilité avec les objets historiques de type `binary` doit être conservée pour ne pas casser les données existantes.
