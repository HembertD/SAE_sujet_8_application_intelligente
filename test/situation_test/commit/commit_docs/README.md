# Calculatrice Python

Une calculatrice simple en ligne de commande écrite en Python.

## Fonctionnalités
- Addition (`add`)
- Soustraction (`sub`)
- Multiplication (`mul`)
- Division (`div`)

## Utilisation

```bash
python calculator.py add 10 5
# Résultat: 15.0

python calculator.py mul 6 7
# Résultat: 42.0

python calculator.py div 10 2
# Résultat: 5.0
```

## Gestion des erreurs
- En cas de division par zéro, un message d'erreur explicite est renvoyé avec un code de sortie non nul.
- Les arguments invalides (non numériques) sont interceptés à la saisie.

## Lancer les tests

```bash
python -m unittest discover tests
```

## Contribution
1. Créez une branche (`git checkout -b feature/ma-fonctionnalite`)
2. Assurez-vous que tous les tests passent avec `python -m unittest discover tests`
3. Rédigez vos commits selon la spécification **Conventional Commits**
4. Ouvrez une Pull Request
