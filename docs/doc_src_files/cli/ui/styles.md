# Moteur de Styles et Séquences ANSI

Fichier source associé : src/cli/ui/styles.go

## Description

Ce module fournit les constantes d'échappement ANSI et les fonctions utilitaires pour le formatage visuel du terminal : couleurs vives, couleurs TrueColor 24-bit, contrôle du curseur et calcul précis des largeurs visuelles.

## Responsabilités

- définir les constantes ANSI fondamentales (`Reset`, `Bold`, `Dim`, `Italic`, `Underline`, `ClearScreen`, `HideCursor`, `ShowCursor`) ;
- fournir une palette de couleurs standard (`Cyan`, `Yellow`, `Green`, `Red`, `Blue`, `Magenta`, `White`, `Gray`) ;
- supporter le rendu TrueColor RGB (`Rgb`, `NeonGradient`) ;
- calculer la largeur visible réelle d'une chaîne contenant des codes d'échappement ANSI (`VisualLen`) via la suppression d'expressions régulières (`StripANSI`) ;
- assurer le remplissage et l'alignement de chaînes (`PadRight`, `PadLeft`, `PadCenter`) sans décaler les bordures des cadres ;
- réinitialiser complètement l'écran et le tampon d'historique de défilement (`ClearTerminal`).

## Points clés

- **Alignement au pixel/caractère près** : l'utilisation naïve de `len(string)` en Go fausse le calcul dès qu'une couleur ANSI ou un caractère Unicode accentué est présent. `VisualLen` compte le nombre exact de runes visibles affichées par le terminal ;
- **Nettoyage absolu** : envoi de la séquence `\033[H\033[2J\033[3J` pour un écran totalement propre à chaque changement de vue.

## Rôle dans l’architecture

Socle de présentation graphique de bas niveau utilisé par tous les composants visuels et toutes les vues du CLI.
