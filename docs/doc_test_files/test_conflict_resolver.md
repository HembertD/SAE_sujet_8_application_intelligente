# Tests de la résolution de conflits de merge

Fichier source associé : `test/test_conflict_resolver.py`

## Description

Ce module couvre la logique de validation et d'interaction du module `conflict_resolver.py`, qui propose une résolution à une section de conflit de merge via Ollama.

## Ce qui est testé

- la génération d'une proposition de résolution valide dès le premier essai ;
- le rejet d'une proposition qui contient encore des marqueurs de conflit (`<<<<<<<`, `=======`, `>>>>>>>`), avec retry puis succès ;
- l'échec (`ConflictResolutionError`) après deux propositions invalides consécutives ;
- le retry sur une réponse non-JSON malgré le format contraint, puis succès (même mécanisme que pour la génération de commit) ;
- le rejet d'un conflit dont les deux côtés (`ours`/`theirs`) sont vides ;
- les règles de validation (`_validate`) isolément de la génération ;
- la délégation au mock (`mock_conflict_resolution`) sans appel réseau quand `MOCK_AI=True`.

## Stratégie de test

Les appels réseau réels sont évités par remplacement de `urllib.request.urlopen` avec un faux objet de réponse, comme pour `test_ollama_llm.py`.

## Points de vigilance

- La validation doit détecter tout marqueur de conflit résiduel : une proposition qui laisse `<<<<<<<`/`=======`/`>>>>>>>` dans le code est pire qu'une absence de proposition, car elle casserait silencieusement la compilation/l'exécution si appliquée sans relecture.
- La proposition n'est jamais appliquée automatiquement par ce module : ça reste la responsabilité de la couche d'affichage (hors périmètre de ces tests).
