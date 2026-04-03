// Package i18n provides internationalization support.
package i18n

import (
	"fmt"
)

// Language represents a supported language.
type Language string

const (
	English    Language = "en"
	French     Language = "fr"
	Chinese    Language = "zh"
	Japanese   Language = "ja"
	Korean     Language = "ko"
	Spanish    Language = "es"
	German     Language = "de"
	Portuguese Language = "pt"
	Italian    Language = "it"
)

// Translator provides translation services.
type Translator struct {
	lang        Language
	messages    map[Language]map[string]string
	pluralizers map[Language]Pluralizer
}

// Pluralizer handles pluralization rules.
type Pluralizer func(n int) string

// NewTranslator creates a new translator.
func NewTranslator(lang Language) *Translator {
	t := &Translator{
		lang:     lang,
		messages: make(map[Language]map[string]string),
	}

	t.loadMessages()
	t.loadPluralizers()

	return t
}

// T translates a message.
func (t *Translator) T(key string, args ...interface{}) string {
	if msgs, ok := t.messages[t.lang]; ok {
		if msg, ok := msgs[key]; ok {
			return fmt.Sprintf(msg, args...)
		}
	}

	// Fallback to English
	if msgs, ok := t.messages[English]; ok {
		if msg, ok := msgs[key]; ok {
			return fmt.Sprintf(msg, args...)
		}
	}

	return key
}

// N translates a pluralized message.
func (t *Translator) N(key string, count int, args ...interface{}) string {
	pluralizer := t.pluralizers[t.lang]
	if pluralizer == nil {
		pluralizer = t.pluralizers[English]
	}

	suffix := pluralizer(count)
	return t.T(key+suffix, append([]interface{}{count}, args...)...)
}

// SetLanguage sets the current language.
func (t *Translator) SetLanguage(lang Language) {
	t.lang = lang
}

func (t *Translator) loadMessages() {
	// English (default)
	t.messages[English] = map[string]string{
		"app.name":          "TokMan",
		"app.description":   "Token-Optimized Command Manager",
		"command.filter":    "Filter command output",
		"command.dashboard": "Launch dashboard",
		"command.config":    "Manage configuration",
		"command.help":      "Show help",
		"savings.total":     "Total tokens saved: %d",
		"savings.command":   "%d tokens saved for this command",
		"error.not_found":   "Command not found: %s",
		"error.invalid":     "Invalid input: %s",
		"status.ok":         "OK",
		"status.error":      "Error",
		"filter.applied":    "Filter %s applied",
		"compression.ratio": "Compression ratio: %.1f%%",
	}

	// French
	t.messages[French] = map[string]string{
		"app.name":          "TokMan",
		"app.description":   "Gestionnaire de Commandes Optimisé pour les Tokens",
		"command.filter":    "Filtrer la sortie de commande",
		"command.dashboard": "Lancer le tableau de bord",
		"command.config":    "Gérer la configuration",
		"command.help":      "Afficher l'aide",
		"savings.total":     "Total de tokens économisés: %d",
		"savings.command":   "%d tokens économisés pour cette commande",
		"error.not_found":   "Commande non trouvée: %s",
		"error.invalid":     "Entrée invalide: %s",
		"status.ok":         "OK",
		"status.error":      "Erreur",
		"filter.applied":    "Filtre %s appliqué",
		"compression.ratio": "Taux de compression: %.1f%%",
	}

	// Chinese
	t.messages[Chinese] = map[string]string{
		"app.name":          "TokMan",
		"app.description":   "令牌优化命令管理器",
		"command.filter":    "过滤命令输出",
		"command.dashboard": "启动仪表板",
		"command.config":    "管理配置",
		"command.help":      "显示帮助",
		"savings.total":     "节省的总令牌数: %d",
		"savings.command":   "此命令节省了 %d 个令牌",
		"error.not_found":   "未找到命令: %s",
		"error.invalid":     "无效输入: %s",
		"status.ok":         "正常",
		"status.error":      "错误",
		"filter.applied":    "已应用过滤器 %s",
		"compression.ratio": "压缩率: %.1f%%",
	}

	// Japanese
	t.messages[Japanese] = map[string]string{
		"app.name":          "TokMan",
		"app.description":   "トークン最適化コマンドマネージャー",
		"command.filter":    "コマンド出力をフィルタリング",
		"command.dashboard": "ダッシュボードを起動",
		"command.config":    "設定を管理",
		"command.help":      "ヘルプを表示",
		"savings.total":     "节省されたトークンの総数: %d",
		"savings.command":   "このコマンドで %d トークンを節約",
		"error.not_found":   "コマンドが見つかりません: %s",
		"error.invalid":     "無効な入力: %s",
		"status.ok":         "OK",
		"status.error":      "エラー",
		"filter.applied":    "フィルター %s を適用しました",
		"compression.ratio": "圧縮率: %.1f%%",
	}

	// Korean
	t.messages[Korean] = map[string]string{
		"app.name":          "TokMan",
		"app.description":   "토큰 최적화 명령 관리자",
		"command.filter":    "명령 출력 필터링",
		"command.dashboard": "대시보드 실행",
		"command.config":    "구성 관리",
		"command.help":      "도움말 표시",
		"savings.total":     "저장된 총 토큰: %d",
		"savings.command":   "이 명령에 대해 %d 토큰 저장",
		"error.not_found":   "명령을 찾을 수 없음: %s",
		"error.invalid":     "잘못된 입력: %s",
		"status.ok":         "정상",
		"status.error":      "오류",
		"filter.applied":    "필터 %s 적용됨",
		"compression.ratio": "압축률: %.1f%%",
	}

	// Spanish
	t.messages[Spanish] = map[string]string{
		"app.name":          "TokMan",
		"app.description":   "Gestor de Comandos Optimizado para Tokens",
		"command.filter":    "Filtrar salida de comando",
		"command.dashboard": "Lanzar panel de control",
		"command.config":    "Gestionar configuración",
		"command.help":      "Mostrar ayuda",
		"savings.total":     "Total de tokens ahorrados: %d",
		"savings.command":   "%d tokens ahorrados para este comando",
		"error.not_found":   "Comando no encontrado: %s",
		"error.invalid":     "Entrada inválida: %s",
		"status.ok":         "OK",
		"status.error":      "Error",
		"filter.applied":    "Filtro %s aplicado",
		"compression.ratio": "Ratio de compresión: %.1f%%",
	}

	// German
	t.messages[German] = map[string]string{
		"app.name":          "TokMan",
		"app.description":   "Token-optimierter Befehlsmanager",
		"command.filter":    "Befehlsausgabe filtern",
		"command.dashboard": "Dashboard starten",
		"command.config":    "Konfiguration verwalten",
		"command.help":      "Hilfe anzeigen",
		"savings.total":     "Gesamte gesparte Tokens: %d",
		"savings.command":   "%d Tokens für diesen Befehl gespart",
		"error.not_found":   "Befehl nicht gefunden: %s",
		"error.invalid":     "Ungültige Eingabe: %s",
		"status.ok":         "OK",
		"status.error":      "Fehler",
		"filter.applied":    "Filter %s angewendet",
		"compression.ratio": "Kompressionsrate: %.1f%%",
	}

	// Portuguese
	t.messages[Portuguese] = map[string]string{
		"app.name":          "TokMan",
		"app.description":   "Gerenciador de Comandos Otimizado para Tokens",
		"command.filter":    "Filtrar saída de comando",
		"command.dashboard": "Iniciar painel",
		"command.config":    "Gerenciar configuração",
		"command.help":      "Mostrar ajuda",
		"savings.total":     "Total de tokens economizados: %d",
		"savings.command":   "%d tokens economizados para este comando",
		"error.not_found":   "Comando não encontrado: %s",
		"error.invalid":     "Entrada inválida: %s",
		"status.ok":         "OK",
		"status.error":      "Erro",
		"filter.applied":    "Filtro %s aplicado",
		"compression.ratio": "Taxa de compressão: %.1f%%",
	}

	// Italian
	t.messages[Italian] = map[string]string{
		"app.name":          "TokMan",
		"app.description":   "Gestore Comandi Ottimizzato per Token",
		"command.filter":    "Filtra output comando",
		"command.dashboard": "Avvia dashboard",
		"command.config":    "Gestisci configurazione",
		"command.help":      "Mostra aiuto",
		"savings.total":     "Totale token risparmiati: %d",
		"savings.command":   "%d token risparmiati per questo comando",
		"error.not_found":   "Comando non trovato: %s",
		"error.invalid":     "Input non valido: %s",
		"status.ok":         "OK",
		"status.error":      "Errore",
		"filter.applied":    "Filtro %s applicato",
		"compression.ratio": "Rapporto di compressione: %.1f%%",
	}
}

func (t *Translator) loadPluralizers() {
	t.pluralizers = map[Language]Pluralizer{
		English: func(n int) string {
			if n == 1 {
				return ""
			}
			return "_plural"
		},
		French: func(n int) string {
			if n <= 1 {
				return ""
			}
			return "_plural"
		},
		Chinese: func(n int) string {
			return "" // No plural in Chinese
		},
		Japanese: func(n int) string {
			return "" // No plural in Japanese
		},
		Korean: func(n int) string {
			return "" // No plural in Korean
		},
		Spanish: func(n int) string {
			if n == 1 {
				return ""
			}
			return "_plural"
		},
		German: func(n int) string {
			if n == 1 {
				return ""
			}
			return "_plural"
		},
		Portuguese: func(n int) string {
			if n == 1 {
				return ""
			}
			return "_plural"
		},
		Italian: func(n int) string {
			if n == 1 {
				return ""
			}
			return "_plural"
		},
	}
}

// GetSupportedLanguages returns list of supported languages.
func GetSupportedLanguages() []Language {
	return []Language{
		English,
		French,
		Chinese,
		Japanese,
		Korean,
		Spanish,
		German,
		Portuguese,
		Italian,
	}
}

// GetLanguageName returns the display name of a language.
func GetLanguageName(lang Language) string {
	names := map[Language]string{
		English:    "English",
		French:     "Français",
		Chinese:    "中文",
		Japanese:   "日本語",
		Korean:     "한국어",
		Spanish:    "Español",
		German:     "Deutsch",
		Portuguese: "Português",
		Italian:    "Italiano",
	}

	if name, ok := names[lang]; ok {
		return name
	}
	return string(lang)
}

// Global translator instance.
var globalTranslator = NewTranslator(English)

// T is a shortcut for global translation.
func T(key string, args ...interface{}) string {
	return globalTranslator.T(key, args...)
}

// SetGlobalLanguage sets the global language.
func SetGlobalLanguage(lang Language) {
	globalTranslator.SetLanguage(lang)
}
