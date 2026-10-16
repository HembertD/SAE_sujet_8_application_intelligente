# Vue Release Notes

Fichier source associé : src/cli/views/release_notes_view.go

## Description

Ce module gère l'écran interactif dédié à la synthèse de version et à la production automatique de changelogs au format Markdown entre deux bornes de révision Git (tags).

## Responsabilités

- inviter l'utilisateur à saisir le tag de départ (par défaut `v0.1.0`) et le tag d'arrivée (par défaut `HEAD`) ;
- déclencher la synthèse des commits auprès du backend Python avec retour visuel animé ;
- afficher le document Markdown catégorisé dans un panneau encadré ;
- proposer l'enregistrement direct du résultat dans un fichier `RELEASE_NOTES.md` à la racine du projet (option masquée et bloquée en cas d'erreur de génération).

## Points clés

- **Catégorisation sémantique** : exploitation de la classification produite par le LLM (Nouveautés, Corrections de bogues, Documentation & Maintenance) ;
- **Export en 1 touche** : écriture immédiate du fichier Markdown sans quitter l'interface ;
- **Boucle de décision robuste** : utilisation de `ReadChoice` pour absorber les appuis vides sur Entrée sans quitter l'écran et verrouillage contextuel de la sauvegarde `[S]`.

## Rôle dans l’architecture

Vue secondaire du sujet 8, concrétisant la seconde fonctionnalité majeure de la SAÉ : la génération automatisée de notes de version.

## Détails déplacés du code source

Anciens commentaires internes de `ShowReleaseNotesFlow` (supprimés du `.go`, conservés ici) : la fonction se structure en trois phases marquées dans le code — saisie des bornes de révision (`fromTag`/`toTag`), génération puis affichage du Markdown dans un panneau encadré, et choix utilisateur final (`S` sauvegarder / `R` retour menu).

## Commentaires de fonction (dupliqués du code source)

- `ShowReleaseNotesFlow()` : Orchestre le flux interactif de génération des release notes.
