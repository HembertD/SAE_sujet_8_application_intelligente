"""
Calculatrice simple en Python.
Fournit les opérations arithmétiques de base en ligne de commande.
"""

from functools import lru_cache
import sys


@lru_cache(maxsize=1024)
def add(a: float, b: float) -> float:
    """Additionne deux nombres avec mise en cache des résultats récurrents."""
    return a + b


@lru_cache(maxsize=1024)
def subtract(a: float, b: float) -> float:
    """Soustrait b de a avec mise en cache."""
    return a - b


@lru_cache(maxsize=1024)
def multiply(a: float, b: float) -> float:
    """Multiplie deux nombres avec mise en cache."""
    return a * b


@lru_cache(maxsize=1024)
def divide(a: float, b: float) -> float:
    """Divise a par b avec mise en cache. Lève ZeroDivisionError si b vaut 0."""
    if b == 0:
        raise ZeroDivisionError("Division par zéro impossible")
    return a / b


# Table de dispatch constante pour éviter les réallocations de dictionnaire à chaque appel
OPERATIONS = {
    "add": add,
    "sub": subtract,
    "mul": multiply,
    "div": divide,
}


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

    if op not in OPERATIONS:
        print(f"Erreur: Opération inconnue '{op}'")
        sys.exit(1)

    try:
        result = OPERATIONS[op](a, b)
        print(f"Résultat: {result}")
    except ZeroDivisionError as e:
        print(f"Erreur: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
