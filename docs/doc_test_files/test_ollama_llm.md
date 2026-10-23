# Tests du client Ollama

Fichier source associé : test/test_ollama_llm.py

## Description

Ce module couvre la logique de validation et d’interaction avec le client Ollama. Il vérifie notamment le comportement du code qui transforme un diff en un message de commit conforme à la norme Conventional Commits.

## Ce qui est testé

- la génération d’un message de commit valide dès le premier essai ;
- le mécanisme de retry lorsqu’une réponse est invalide ;
- l’échec après plusieurs réponses invalides ;
- les règles de validation du message (`type`, `scope`, `subject`) ;
- la vérification de joignabilité rapide et extraction des modèles via `check_ollama_reachable()` / `ping_ollama()` (succès, échec réseau, mesure de latence, récupération des modèles avec `fetch_models`, sérialisation `to_dict`, compatibilité tuple) ;
- la compatibilité avec les anciennes structures de données binaires ;
- l'override déterministe du type via `feedback` : une consigne contenant une formulation d'intention explicite ("c'est un fix, je veux un fix") force le type final même si le modèle en propose un autre ; une consigne sans mot-clé de type, ou avec deux types contradictoires, ne déclenche aucun override ; la simple présence du mot "test" (homonyme d'un type et d'un mot français courant) sans formulation d'intention ne doit pas non plus déclencher d'override ; `_extract_forced_type()` est aussi testée directement, isolément de la génération ;
- le retry sur réponse non-JSON (`MalformedResponseError`) : une première réponse dont le `content` n'est pas du JSON déclenche le même mécanisme de retry que pour une réponse invalide côté règles métier ; si les deux essais échouent avec du contenu non-JSON, `CommitMessageValidationError` est levée au lieu de propager l'erreur brute.

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
- La détection d'un type imposé par `feedback` doit rester prudente : exiger une formulation d'intention explicite, pas la simple présence du mot, sinon toute mention anodine d'un mot comme "test" casserait le type choisi par le modèle.
