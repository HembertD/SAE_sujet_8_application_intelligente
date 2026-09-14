"""
Désactive l'environnement de test en masquant tous les éléments Git et sensibles (ajout du suffixe '_test').
Masque :
- .git -> .git_test
- .gitignore -> .gitignore_test
- .github -> .github_test
- .gitattributes -> .gitattributes_test
- .gitmodules -> .gitmodules_test
- .env -> .env_test

Compatible multi-OS (Linux, macOS, Windows) via pathlib.
"""

from pathlib import Path
import shutil
from typing import List, Tuple, Optional

# Éléments ayant un impact avec Git ou sécurité (.env) à masquer
GIT_TARGETS = [".git", ".gitignore", ".github", ".gitattributes", ".gitmodules", ".env"]


def deinit_test(base_dir: Optional[Path] = None) -> List[Tuple[Path, Path]]:
    """
    Masque tous les fichiers et dossiers ayant un impact Git ou sécurité (* -> *_test) dans base_dir.
    Ne produit aucun affichage dans la console.

    :param base_dir: Répertoire racine de recherche (par défaut le dossier du script).
    :return: Liste de tuples (chemin_source, chemin_cible) pour chaque élément renommé.
    """
    if base_dir is None:
        base_dir = Path(__file__).resolve().parent
    else:
        base_dir = Path(base_dir).resolve()

    renamed: List[Tuple[Path, Path]] = []

    for name in GIT_TARGETS:
        items = [p for p in base_dir.rglob(name)]
        # Traitement du plus profond au plus superficiel
        items.sort(key=lambda p: len(p.parts), reverse=True)

        for item in items:
            target = item.with_name(f"{name}_test")
            if target.is_dir():
                shutil.rmtree(target)
            elif target.exists():
                target.unlink()
            item.rename(target)
            renamed.append((item, target))

    return renamed


if __name__ == "__main__":
    print("=== Désactivation de l'environnement de test (actifs -> *_test) ===")
    results = deinit_test()
    if not results:
        print("ℹ️  Aucun élément actif trouvé.")
    else:
        for src, dst in results:
            print(f"✔ Masqué : {src.name} -> {dst.name} ({dst.parent.name})")
        print(f"🔒 {len(results)} élément(s) neutralisé(s) en *_test avec succès !")
