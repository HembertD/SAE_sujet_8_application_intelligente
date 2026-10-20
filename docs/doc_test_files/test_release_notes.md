# Tests de la génération de release notes

Fichier source associé : `test/test_release_notes.py`

## Description

Ce module couvre la logique de validation et d'interaction du module `release_notes.py`, qui génère le Markdown des release notes entre deux tags via Ollama.

## Ce qui est testé

- la génération d'un Markdown valide dès le premier essai ;
- le rejet d'une réponse au `markdown` vide, avec retry puis succès ;
- l'échec (`ReleaseNotesError`) après deux réponses invalides consécutives ;
- le retry sur une réponse non-JSON malgré le format contraint, puis succès (même mécanisme que pour la génération de commit et la résolution de conflit) ;
- le rejet d'une liste de commits vide ;
- les règles de validation (`_validate`) isolément de la génération ;
- la délégation au mock (`mock_release_notes`) sans appel réseau quand `MOCK_AI=True`.

## Stratégie de test

Les appels réseau réels sont évités par remplacement de `urllib.request.urlopen` avec un faux objet de réponse, comme pour `test_ollama_llm.py` et `test_conflict_resolver.py`. Les commits de test sont de simples `SimpleNamespace(sha=..., message=...)` plutôt que le type réel du `GitWrapper`, pour ne pas créer de dépendance de test vers un autre module.

## Points de vigilance

- Les tests `service/test_core.py::test_action_release_notes_*` patchent le même `urllib.request.urlopen` global : comme `generate_release_notes()` appelle en interne `llm.call_chat_json()`, qui passe par ce même symbole, ces tests restent valides sans modification malgré le remplacement de l'appel HTTP ad-hoc dans `core.py`.
