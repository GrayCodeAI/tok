<div align="center">

# 🚀 TokMan: 60-90% Token-Einsparung für KI-Codierassistenten

**Token-bewusster CLI-Proxy**

31-Schichten-Komprimierungspipeline für 60-90% Token-Ersparnis

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Discord](https://img.shields.io/discord/1470188214710046894?label=Discord&logo=discord&style=flat-square&color=5865F2)](https://discord.gg/HrVA7ePyV)

[🌐 English](README.md) · [🇫🇷 Français](README_fr.md) · [🇨🇳 中文](README_zh.md) · [🇯🇵 日本語](README_ja.md) · [🇪🇸 Español](README_es.md) · **[🇩🇪 Deutsch](README_de.md)** · [🇰🇷 한국어](README_ko.md)

</div>

---

## 💡 Was ist TokMan?

TokMan fängt CLI-Befehle ab und wendet eine intelligente **31-Schichten-Komprimierungspipeline** an, um die Token-Nutzung für KI-Codierassistenten drastisch zu reduzieren.

## ✨ Hauptfunktionen

- **60-90% Token-Reduzierung**
- **31 Schichten** forschungsbasierte Komprimierung
- **RewindStore** - Archiv ohne Datenverlust
- **Lernmodus** - Automatische Mustererkennung
- **YAML-Filter** - Einfache Filterdefinitionen
- **Sitzungswiederherstellung** - Neustart nach Absturz
- **MCP-Integration** - Kompatibel mit Claude Desktop, Cursor

## 🚀 Schnellstart

```bash
# Installieren
brew install GrayCodeAI/tap/tokman

# Oder via Go
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Für Claude Code einrichten
tokman init -g

# Überprüfen
tokman doctor
```

## 📊 Reale Einsparungen

| Befehl | Vorher | Nachher | Ersparnis |
|--------|--------|---------|-----------|
| `go test ./...` | 689 Tokens | 16 Tokens | **97,7%** |
| `git status` | 112 Tokens | 16 Tokens | **85,7%** |
| `cargo test` | 591 Tokens | 5 Tokens | **99,2%** |

---

<div align="center">

**⭐ Gib uns einen Stern, wenn TokMan dir hilft Tokens zu sparen!**

</div>
