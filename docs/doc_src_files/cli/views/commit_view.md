# Vue Génération de Commit

Fichier source associé : src/cli/views/commit_view.go

## Description

Ce module orchestre le workflow interactif de génération, de validation et d'application des messages de commit. Il structure le dialogue entre l'utilisateur et l'IA à travers une boucle de décision complète.

## Workflow en 4 étapes

1. **Analyse du diff** : demande au backend Python la liste des fichiers modifiés et les affiche. Si aucun fichier n'est indexé, propose une indexation automatique (`git add -A`) via Python.
2. **Génération par le LLM** : déclenche le spinner d'attente pendant que Python interroge le modèle distant.
3. **Restitution structurée** : affiche la carte du message proposé au format Conventional Commits.
4. **Boucle de décision utilisateur** :
   - `[V]` **Valider** : applique le commit via Git et propose d'enchaîner sur un `git push` ;
   - `[R]` **Régénérer avec consigne** : demande à l'utilisateur un texte de feedback (ex: "sois plus concis", "c'est un fix") et relance l'inférence LLM avec ce contexte ;
   - `[M]` **Modifier manuellement** : permet d'ajuster le type, le scope, le sujet ou le corps directement dans le terminal avant validation ;
   - `[A]` **Annuler** : abandonne l'opération sans impacter le dépôt Git.

## Points clés

- **Respect absolu du rôle front-end** : aucune commande Git directe n'est exécutée depuis ce fichier ; toutes les requêtes (`stage`, `commit`, `push`, `llm`) sont transmises au `BackendClient` ;
- **Contrôle utilisateur total** : l'utilisateur reste maître absolu de la validation finale du commit.

## Rôle dans l’architecture

Vue maîtresse du sujet 8, réalisant l'automatisation de la rédaction des messages de commit assistée par intelligence artificielle.
