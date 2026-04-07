<div align="center">

# 🚀 TokMan: Réduisez l'Usage des Tokens de 60-90%

**Proxy CLI conscient des tokens pour assistants de codage IA**

Pipeline de compression à 31 couches atteignant 60-90% d'économie de tokens

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Discord](https://img.shields.io/discord/1470188214710046894?label=Discord&logo=discord&style=flat-square&color=5865F2)](https://discord.gg/HrVA7ePyV)

[🌐 Anglais](README.md) · **[🇫🇷 Français](README_fr.md)** · [🇨🇳 中文](README_zh.md) · [🇯🇵 日本語](README_ja.md) · [🇪🇸 Español](README_es.md) · [🇩🇪 Deutsch](README_de.md) · [🇰🇷 한국어](README_ko.md)

</div>

---

## 💡 Qu'est-ce que TokMan ?

TokMan intercepte les commandes CLI et applique un pipeline de compression intelligent à **31 couches** pour réduire drastiquement l'usage de tokens pour les assistants de codage IA.

```
┌──────────────────────────────────────────────────────────────┐
│  Entrée: 10 000 tokens  →  Pipeline TokMan  →  Sortie: 1 500 │
│                                                                │
│  💰 Économie: $0.085 → $0.013  (réduction de 85%)            │
│  ⚡ Rapidité:  Réponses IA plus rapides                       │
│  🎯 Qualité:   Préserve les informations critiques            │
└──────────────────────────────────────────────────────────────┘
```

## ✨ Fonctionnalités Clés

- **60-90% de réduction** de tokens
- **31 couches** de compression basées sur la recherche
- **RewindStore** - Archivage zéro-perte
- **Mode d'apprentissage** - Découverte automatique de motifs
- **Filtres YAML** - Définissez vos filtres simplement
- **Récupération de session** - Reprendre après un crash
- **Intégration MCP** - Compatible Claude Desktop, Cursor

## 🚀 Démarrage Rapide

```bash
# Installer
brew install GrayCodeAI/tap/tokman

# Ou via Go
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Configurer pour Claude Code
tokman init -g

# Vérifier
tokman doctor
```

## 📊 Impact Réel

| Commande | Avant | Après | Économie |
|----------|-------|-------|----------|
| `go test ./...` | 689 tokens | 16 tokens | **97.7%** |
| `git status` | 112 tokens | 16 tokens | **85.7%** |
| `cargo test` | 591 tokens | 5 tokens | **99.2%** |

---

<div align="center">

**⭐ Mettez une étoile si TokMan vous aide à économiser des tokens !**

</div>
