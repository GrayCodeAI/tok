# tok

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **Écrivez moins, obtenez plus.** Compressez vos prompts avant de les envoyer. Filtrez les sorties bruyantes pour une meilleure lisibilité.

---

## Qu'est-ce que tok ?

tok est une CLI d'optimisation unifiée de tokens conçue pour les assistants de codage IA :

1. **Compression de prompts** — Réduit les tokens d'entrée envoyés à l'IA
2. **Filtrage de sortie terminal** — Affiche uniquement l'essentiel
3. **Configuration du ton de l'agent IA** — Fait répondre les agents de manière concise, économisant des tokens de sortie

## Installation

```bash
go install github.com/GrayCodeAI/tok/cmd/tok@latest
```

Ou avec Homebrew :

```bash
brew install lakshmanpatel/tap/tok
```

## Démarrage rapide

```bash
# Compresser un prompt
$ tok compress -mode ultra -input "Peux-tu m'aider à comprendre pourquoi ce composant React se ré-affiche à chaque changement de props ?"
Composant React re-renders sur changement de props. Pourquoi ?

# Filtrer la sortie d'une commande
$ tok npm test
# 200 lignes de résultats → 3 lignes : succès/échec + échecs

# Configurer le ton de l'agent
$ tok on ultra       # Breveté maximale
$ tok status         # Voir le mode actuel
```

## Modes de compression

| Mode | Style | Économie d'entrée |
|------|-------|------------------|
| `lite` | Supprime le remplissage, garde la grammaire | ~20% |
| `full` | Supprime les articles, fragments OK | ~40% _(par défaut)_ |
| `ultra` | Télégraphique, abréviations | ~60% |

## Documentation

Consultez le [README en anglais](README.md) pour la documentation complète.

## Licence

[MIT](LICENSE)
