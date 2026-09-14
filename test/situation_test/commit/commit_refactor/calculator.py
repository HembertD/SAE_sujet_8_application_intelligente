"""
Calculatrice simple en Python.
Fournit les opérations arithmétiques de base en ligne de commande.
"""

import sys
from typing import Callable, Dict


class Calculator:
    """Classe encapsulant les opérations arithmétiques de la calculatrice."""

    def __init__(self) -> None:
        self._operations: Dict[str, Callable[[float, float], float]] = {
            "add": self.add,
            "sub": self.subtract,
            "mul": self.multiply,
            "div": self.divide,
        }

    @staticmethod
    def add(a: float, b: float) -> float:
        """Additionne deux nombres."""
        return a + b

    @staticmethod
    def subtract(a: float, b: float) -> float:
        """Soustrait b de a."""
        return a - b

    @staticmethod
    def multiply(a: float, b: float) -> float:
        """Multiplie deux nombres."""
        return a * b

    @staticmethod
    def divide(a: float, b: float) -> float:
        """Divise a par b. Lève ZeroDivisionError si b vaut 0."""
        if b == 0:
            raise ZeroDivisionError("Division par zéro impossible")
        return a / b

    def execute(self, op: str, a: float, b: float) -> float:
        """Exécute l'opération demandée."""
        if op not in self._operations:
            raise ValueError(f"Opération inconnue '{op}'")
        return self._operations[op](a, b)


# Fonctions wrapper pour la rétrocompatibilité
_default_calc = Calculator()
add = _default_calc.add
subtract = _default_calc.subtract
multiply = _default_calc.multiply
divide = _default_calc.divide


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

    calc = Calculator()
    try:
        result = calc.execute(op, a, b)
        print(f"Résultat: {result}")
    except (ValueError, ZeroDivisionError) as e:
        print(f"Erreur: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
