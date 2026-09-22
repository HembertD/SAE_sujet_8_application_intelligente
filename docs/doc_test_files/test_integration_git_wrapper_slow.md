# Test d'intégration réel avec le wrapper Git

Fichier source associé : test/test_integration_git_wrapper_slow.py

## Description

Ce test vérifie que le `list[DiffFile]` réellement produit par `GitWrapper.get_staged_diff()` (module Git) est directement consommable par `generate_commit_message()` (module LLM), sans conversion manuelle entre les deux. Il s'appuie sur les 12 vrais dépôts Git de `test/situation_test/commit/` plutôt que sur des fixtures écrites à la main.

## Ce qui est testé

- que le wrapper Git et le client LLM communiquent sans erreur de compatibilité de contrat sur 8 scénarios réels (feat, fix, docs, refactor, test, chore, perf, build) ;
- que chaque scénario produit un `CommitMessage` avec un type et un subject non vides.

## Stratégie de test

Avant chaque scénario, le test indexe (`git add -A`) les modifications locales du dépôt factice correspondant, puisque `get_staged_diff()` ne lit que le diff indexé (`git diff --cached`). Nécessite `GitPython` installé, les fixtures activées (`python test/situation_test/init_test.py`), et un accès réseau au serveur Ollama partagé. Marqué `slow`, exclu par défaut. Compte 5 à 10 minutes pour les 8 scénarios selon la charge du serveur.

## Cas importants

### compatibilité de contrat entre deux modules

Le test ne compare pas d'égalité stricte de type : son but premier est de prouver qu'aucune erreur (attribut manquant, format inattendu) ne survient quand un vrai `DiffFile` du module Git traverse le module LLM.

## Points de vigilance

- Les dépôts de `situation_test/` contiennent plusieurs fichiers modifiés à la fois (code, tests, CI, doc), donc des diffs plus volumineux que les golden files : latence et probabilité de timeout plus élevées, surtout si le serveur partagé est chargé.
- Toujours désactiver les fixtures après usage (`python test/situation_test/deinit_test.py`) pour ne pas laisser de vrais `.git` actifs dans l'arborescence suivie par le dépôt principal.
