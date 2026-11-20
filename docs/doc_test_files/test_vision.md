# Tests de la description d'image par un modèle de vision

Fichier source associé : `test/test_vision.py`

## Description

Ce module couvre la logique de validation et d'interaction du module `vision.py`, qui décrit le contenu d'une image ajoutée/supprimée ou le changement visuel entre deux versions d'une image modifiée, via le modèle de vision Ollama.

## Ce qui est testé

- la sélection des images envoyées selon `status` : une seule image pour un ajout (`"A"`) ou une suppression (`"D"`), les deux images (avant/après) pour une modification ;
- le libellé d'intention transmis au modèle selon ces mêmes cas (`"ajoutée"`, `"supprimée"`, `"avant"`/`"après"`) ;
- le retry sur description vide, puis succès ;
- l'échec (`ImageDescriptionError`) après deux réponses invalides consécutives ;
- le retry sur réponse non-JSON malgré le format contraint, puis succès (même mécanisme que pour la génération de commit, de release notes et la résolution de conflit) ;
- le rejet d'un appel sans aucune image (`new_image` et `old_image` tous les deux `None`) ;
- les règles de validation (`_validate`) isolément de la génération ;
- la délégation au mock (`mock_describe_image`) sans appel réseau quand `MOCK_AI=True`.

## Stratégie de test

Les appels réseau réels sont évités par remplacement de `urllib.request.urlopen` avec un faux objet de réponse, comme pour `test_ollama_llm.py`. Les tests inspectent aussi le corps de la requête envoyée (`request.data`) pour vérifier que le nombre d'images et le texte d'intention correspondent bien au `status` passé en argument.

## Points de vigilance

- Le mauvais choix d'images pour un `status` donné (ex: envoyer les deux images pour une suppression) tromperait le modèle sur ce qu'il doit décrire — les tests vérifient explicitement le nombre d'images envoyées pour chaque cas.
- Vérifié également en conditions réelles (hors suite automatisée) contre le serveur Ollama partagé avec le modèle `qwen3-vl:8b-instruct` : une comparaison de deux images (fond blanc → fond gris) a produit une description textuelle correcte du changement.
