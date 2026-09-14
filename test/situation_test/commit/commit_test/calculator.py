"""
Calculatrice simple en Python.
Fournit les opérations arithmétiques de base en ligne de commande.
"""

import sys


def add(a: float, b: float) -> float:
    """Additionne deux nombres."""
    return a + b


def subtract(a: float, b: float) -> float:
    """Soustrait b de a."""
    return a - b


def multiply(a: float, b: float) -> float:
    """Multiplie deux nombres."""
    return a * b


def divide(a: float, b: float) -> float:
    """Divise a par b. Lève ZeroDivisionError si b vaut 0."""
    if b == 0:
        raise ZeroDivisionError("Division par zéro impossible")
    return a / b


def main():
    if len(sys.argv) != 4:
        print("Usage: python calculator.py <operation> <a> <b>")
        print("Opérations supportées: add, sub, mul, div")
        sys.exit(1)

    op = sys.argv[1]
    try:
        a = float(sys.argv[2])
        b = float(sys.argv[3])
    except ValueError:
        print("Erreur: Les arguments doivent être des nombres valides.")
        sys.exit(1)

    operations = {
        "add": add,
        "sub": subtract,
        "mul": multiply,
        "div": divide,
    }

    if op not in operations:
        print(f"Erreur: Opération inconnue '{op}'")
        sys.exit(1)

    try:
        result = operations[op](a, b)
        print(f"Résultat: {result}")
    except ZeroDivisionError as e:
        print(f"Erreur: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
