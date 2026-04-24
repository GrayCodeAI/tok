# tok

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **Escribe menos, obtén más.** Comprime tus prompts antes de enviarlos. Filtra la salida ruidosa para mejorar la legibilidad.

---

## ¿Qué es tok?

tok es una CLI de optimización unificada de tokens diseñada para asistentes de codificación con IA:

1. **Compresión de prompts** — Reduce los tokens de entrada enviados a la IA
2. **Filtrado de salida de terminal** — Muestra solo lo que importa
3. **Configuración de tono del agente IA** — Hace que los agentes respondan de forma concisa, ahorrando tokens de salida

## Instalación

```bash
go install github.com/GrayCodeAI/tok/cmd/tok@latest
```

O con Homebrew:

```bash
brew install GrayCodeAI/tap/tok
```

## Inicio rápido

```bash
# Comprimir un prompt
$ tok compress -mode ultra -input "Por favor, implementa un sistema de autenticación de usuarios con tokens JWT"
Implementar auth de usuario con JWT.

# Filtrar salida de comandos
$ tok npm test
# 200 líneas de resultados → 3 líneas: aprobado/fallido + fallos

# Configurar tono del agente
$ tok on ultra       # Máxima brevedad
$ tok status         # Ver modo actual
```

## Modos de compresión

| Modo | Estilo | Ahorro de entrada |
|------|--------|------------------|
| `lite` | Elimina relleno, mantiene gramática | ~20% |
| `full` | Elimina artículos, fragmentos OK | ~40% _(por defecto)_ |
| `ultra` | Telegráfico, abreviaturas | ~60% |

## Documentación

Consulta el [README en inglés](README.md) para la documentación completa.

## Licencia

[MIT](LICENSE)
