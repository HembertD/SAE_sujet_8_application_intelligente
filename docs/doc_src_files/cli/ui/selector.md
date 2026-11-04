# Sélecteur Interactif de Commits (Selector)

Fichier source associé : src/cli/ui/selector.go

## Description

Le composant `SelectCommit` fournit un sélecteur TUI interactif en plein écran avec navigation au clavier et défilement fenêtré pour naviguer dans l'historique complet des commits Git.

## Responsabilités

- basculer le terminal en mode non canonique sans écho (`stty -echo -icanon min 1`) pour intercepter les frappes clavier immédiatement ;
- restaurer systématiquement l'état initial du terminal et la visibilité du curseur à la sortie ;
- gérer une fenêtre déroulante (viewport de 10 lignes par défaut) avec indicateurs `▲` et `▼` pour naviguer élégamment dans de grands historiques ;
- surligner le commit actuellement sous le curseur (`❯`, fond et couleurs distinctes) avec son SHA, tag éventuel, sujet tronqué et date ;
- intercepter les touches de navigation : flèches Haut/Bas, `j`/`k`, Page Up/Down, `g`/`G` (Home/End), Entrée pour valider, `q`/Échap pour annuler ;
- assurer un repli fluide sans blocage lorsque l'environnement n'est pas un terminal interactif (tests automatisés, redirection de flux).

## Rôle dans l’architecture

Composant d'interface utilisateur clé pour l'ergonomie de la génération de release notes, évitant la saisie manuelle de hashes ou de révisions Git complexes.
