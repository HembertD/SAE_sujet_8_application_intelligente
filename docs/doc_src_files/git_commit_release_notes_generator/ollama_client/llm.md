# Client LLM

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/llm.py

## Description

Ce module gère la génération du message de commit via Ollama. Il construit le résumé du diff, envoie une requête structurée à l’API de chat, valide la réponse JSON et la transforme en objet métier.

## Flux de travail

1. Rejeter immédiatement une liste de fichiers vide (`ValueError`), sans appel réseau.
2. Lire les fichiers différences et les convertir en résumé exploitable.
3. Construire le prompt système et le message utilisateur.
4. Appeler Ollama via une requête JSON structurée, avec une fenêtre de contexte (`num_ctx`) fixée explicitement.
5. Vérifier que la réponse est bien un JSON conforme.
6. En cas d’erreur, relancer une seconde fois avec le message invalidé en contexte.
7. Convertir la réponse validée en objet CommitMessage.

## Points clés

- format : le modèle répond avec un schéma JSON strict ;
- validation : le code refuse les messages trop longs, absents ou non conformes ;
- robustesse : un retry est tenté pour corriger une réponse invalide — y compris si le contenu renvoyé n'est carrément pas du JSON (`MalformedResponseError`), pas seulement en cas d'échec des règles métier (`_validate`) ;
- fenêtre de contexte : `options.num_ctx` est fixé explicitement dans la requête (`OLLAMA_NUM_CTX`, config.py) car Ollama utilise sinon une fenêtre par défaut bien plus petite que celle annoncée pour le modèle, ce qui tronquait silencieusement les diffs volumineux ;
- disponibilité : `check_ollama_reachable()` permet de vérifier rapidement (quelques secondes) si le serveur répond, avant de lancer une génération qui pourrait sinon attendre le timeout complet ;
- simulation hors-ligne : si `config.MOCK_AI` est activé, `generate_commit_message` et `check_ollama_reachable` délèguent immédiatement au module `mock_llm` sans aucune requête réseau ;
- override déterministe du type : si le `feedback` utilisateur contient une formulation d'intention explicite ("c'est un fix", "je veux un feat", "type: chore") suivie d'un seul type Conventional Commits connu, `_extract_forced_type()` le détecte et `_apply_forced_type()` l'impose sur le `CommitMessage` final, quel que soit ce que le modèle a réellement répondu pour ce champ. Le modèle garde la main sur `subject`/`body`/`scope`. Si aucun type n'est cité, ou si plusieurs types contradictoires le sont, aucun override n'est appliqué. Une formulation d'intention est exigée (pas la simple présence du mot) car certains types (`test`) sont aussi des mots français courants.

## Fonctions exposées

- `generate_commit_message(diff_files, feedback="")` : fonction principale, décrite ci-dessus. Le `feedback` sert à la fois de consigne de révision injectée dans le prompt et, si un type y est explicitement cité, de source d'override déterministe du type final.
- `check_ollama_reachable(timeout=3.0, fetch_models=False)` (alias `ping_ollama`) : fonction unique et centralisée de diagnostic Ollama. Vérifie la joignabilité (`/api/version`), mesure la latence réseau en millisecondes et extrait optionnellement les modèles installés (`/api/tags` si `fetch_models=True`). Retourne un objet `PingResult(reachable, error, latency_ms, installed_models)` déballable directement en tuple `(reachable, error)` ou sérialisable via `.to_dict()`. Utilisée de manière transverse par `action_generate_commit`, `action_release_notes` et `action_ping_ollama` (écran de configuration Go).

## Dépendances

- [config.md](../config.md)
- [models.md](../models.md)
- [mock_llm.md](mock_llm.md)
- [exceptions.md](exceptions.md)
- [commit_message_system.md](prompts/commit_message_system.md)

## Détails déplacés du code source

- `PingResult` : outre l'accès par attributs (`.reachable`, `.latency_ms`, `.error`, `.installed_models`), il supporte le déballage en tuple (`reachable, error = check_ollama_reachable()`) via `__iter__`/`__getitem__`, et la comparaison directe à un tuple de 2 éléments via `__eq__`. `to_dict()` renvoie `success`, `reachable`, `latency_ms`, `installed_models`, et n'ajoute la clé `error` que si elle est non vide.
- `check_ollama_reachable()` : ne lève jamais d'exception ; mesure la latence via `/api/version`, et n'interroge `/api/tags` que si `fetch_models=True` (échec de récupération des modèles seulement tracé en `debug`).
- `generate_commit_message()` : un seul retry est tenté, pour deux causes possibles — échec de `_validate()` (réponse JSON mais non conforme aux règles métier) ou `MalformedResponseError` (réponse pas du tout JSON malgré le format contraint). Dans les deux cas, la réponse invalide est réinjectée comme message `assistant` (JSON sérialisé, ou contenu brut si non-JSON), suivie d'un message `user` signalant l'erreur constatée et exigeant un JSON conforme. Si le second essai échoue aussi (pour l'une ou l'autre des deux causes), lève `CommitMessageValidationError` plutôt que de renvoyer un message bancal.
- `_build_diff_summary()` : assemble un résumé textuel `Fichier: {path} (statut={status}, +{added}/-{removed})` suivi du `patch`, fichiers joints par une double newline.
- `_validate(raw)` : retourne `None` si valide, sinon un message explicite (`type` hors liste autorisée, `subject` manquant ou vide, `subject` de plus de 72 caractères, `subject` se terminant par un point).
- `_to_commit_message(raw)` : mappe `type`, `scope`, `subject`, `body` du dictionnaire validé vers `CommitMessage`.
- `_extract_forced_type(feedback)` : détecte si le feedback impose explicitement un type. Un préfixe d'intention explicite est requis avant le mot-clé de type (`c'est`, `je veux`, `mets en`, `type:`, ...) sinon un mot comme `test` matcherait n'importe quelle phrase française normale sans rapport avec une consigne. Si un seul type connu est trouvé, il est appliqué de façon déterministe ; si aucun ou plusieurs types contradictoires sont trouvés, retourne `None` et le modèle garde la main.
- `_apply_forced_type(msg, forced_type)` : écrase `msg.type` par le type imposé (avec trace `info` du type proposé par le modèle) ; appliqué aux trois sorties de `generate_commit_message` (mock, premier essai, retry).

## Commentaires de fonction (dupliqués du code source)

- `_extract_forced_type()` : Détecte le type imposé explicitement par le retour utilisateur.
- `_apply_forced_type()` : Applique le type imposé au message généré si nécessaire.
- `PingResult.__iter__()` : Permet le déballage du résultat en tuple.
- `PingResult.__getitem__()` : Permet l'accès au résultat par index.
- `PingResult.__eq__()` : Compare le résultat à un tuple ou un objet.
- `PingResult.to_dict()` : Sérialise le résultat du diagnostic en dictionnaire.
- `check_ollama_reachable()` : Vérifie la joignabilité du serveur Ollama et mesure la latence.
- `generate_commit_message()` : Génère un message de commit à partir des diffs stagés.
- `_build_diff_summary()` : Assemble le résumé textuel des diffs pour le prompt.
- `_call_chat()` : Envoie les messages au chat Ollama et renvoie le JSON.
- `_validate()` : Valide la réponse brute et signale l'erreur éventuelle.
- `_to_commit_message()` : Convertit la réponse validée en objet CommitMessage.
