# TokMan IntelliJ Plugin

A JetBrains IDE plugin for TokMan token optimization integration.

## Status: Skeleton

This is a placeholder for the JetBrains IDE plugin. Full implementation requires:
- IntelliJ Platform SDK setup
- Gradle build configuration
- Plugin.xml configuration

## Quick Start (Manual)

```bash
# Install TokMan
go install github.com/GrayCodeAI/tokman@latest

# For JetBrains IDEs, the plugin will:
# 1. Detect terminal commands
# 2. Run them through tokman
# 3. Display compressed output
```

## Features Planned

- [ ] Terminal output interception
- [ ] Run configuration integration
- [ ] Tool window for statistics
- [ ] Settings panel
- [ ] Run configurations for tokman commands

## Development Setup

1. Download IntelliJ IDEA Community Edition
2. Install IntelliJ Platform SDK
3. Clone this repository
4. Import as Gradle project
5. Run/debug configuration

## Project Structure

```
jetbrains/
├── src/
│   └── main/
│       ├── java/          # Kotlin/Java source
│       ├── resources/     # Plugin resources
│       └── kotlin/        # Kotlin code
├── build.gradle.kts       # Gradle build config
└── settings.gradle.kts    # Gradle settings
```

## Reference

- [IntelliJ Platform SDK](https://jetbrains.org/intellij/sdk/docs/)
- [Plugin Development Guide](https://jetbrains.org/intellij/sdk/docs/basics.html)
- [GitHub Actions Plugin](https://github.com/cthagost/idea-github-actions) - similar reference plugin