# Script de mesure de conformité et de latence

Fichier source associé : scripts/run_metrics.py

## Description

Outil de développement (pas le module lui-même) qui fait tourner `generate_commit_message()` sur les 6 golden files et affiche un tableau récapitulatif : conformité, correspondance du type attendu, taux de retry, latence.

## Flux de travail

1. Ajoute un handler de logging temporaire pour compter les warnings de retry émis par `llm.py`, sans modifier son API publique.
2. Pour chaque golden file : chronomètre l'appel à `generate_commit_message()`, capture le résultat ou l'échec de validation.
3. Affiche un tableau ligne par ligne (scénario, résultat, type attendu/obtenu, retry, latence) puis un récapitulatif chiffré.

## Points clés

- écrit volontairement sur stdout : ce n'est pas le module de production, qui lui doit rester silencieux ;
- nécessite un accès réseau au serveur Ollama partagé, aucun mock ;
- sert de base chiffrée pour la soutenance (taux de conformité, latence moyenne).

## Dépendances

- [llm.md](../doc_src_files/git_commit_release_notes_generator/ollama_client/llm.md)
- [golden_diffs.md](../doc_test_files/fixtures/golden_diffs.md)

## Détails déplacés du code source

- Usage : `python scripts/run_metrics.py`. `_RetryCounter` compte les `warning` contenant `invalide` émis par `llm.py` pour détecter les générations ayant nécessité un retry, sans toucher à son API publique.
- `_print_report()` affiche le tableau (`Scénario`, `Résultat`, `Attendu`, `Obtenu`, `Retry`, `Latence`) puis les taux de conformité, de type exact, de retry et les latences moyenne/min/max.

## Commentaires de fonction (dupliqués du code source)

- `_RetryCounter.__init__()` : Initialise le compteur de retries détectés.
- `_RetryCounter.emit()` : Compte les avertissements signalant une génération invalide.
- `main()` : Mesure la conformité et la latence sur les diffs de référence.
- `_print_report()` : Affiche le tableau récapitulatif des mesures.
