<div align="center">

# 🚀 TokMan: Reducción del Uso de Tokens en 60-90% para Asistentes IA

**Proxy CLI consciente de tokens**

Pipeline de compresión de 31 capas logrando 60-90% de ahorro

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Discord](https://img.shields.io/discord/1470188214710046894?label=Discord&logo=discord&style=flat-square&color=5865F2)](https://discord.gg/HrVA7ePyV)

[🌐 English](README.md) · [🇫🇷 Français](README_fr.md) · [🇨🇳 中文](README_zh.md) · [🇯🇵 日本語](README_ja.md) · **[🇪🇸 Español](README_es.md)** · [🇩🇪 Deutsch](README_de.md) · [🇰🇷 한국어](README_ko.md)

</div>

---

## 💡 ¿Qué es TokMan?

TokMan intercepta comandos CLI y aplica un pipeline de compresión inteligente de **31 capas** para reducir drásticamente el uso de tokens para asistentes de codificación IA.

## ✨ Funcionalidades Clave

- **60-90% de reducción** de tokens
- **31 capas** de compresión basadas en investigación
- **RewindStore** - Archivo sin pérdida
- **Modo aprendizaje** - Descubrimiento automático de patrones
- **Filtros YAML** - Define tus filtros fácilmente
- **Recuperación de sesión** - Reanudar tras un fallo
- **Integración MCP** - Compatible con Claude Desktop, Cursor

## 🚀 Inicio Rápido

```bash
# Instalar
brew install GrayCodeAI/tap/tokman

# O con Go
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Configurar para Claude Code
tokman init -g

# Verificar
tokman doctor
```

## 📊 Ahorro Real

| Comando | Antes | Después | Ahorro |
|---------|-------|---------|--------|
| `go test ./...` | 689 tokens | 16 tokens | **97.7%** |
| `git status` | 112 tokens | 16 tokens | **85.7%** |
| `cargo test` | 591 tokens | 5 tokens | **99.2%** |

---

<div align="center">

**⭐ ¡Danos una estrella si TokMan te ayuda a ahorrar tokens!**

</div>
