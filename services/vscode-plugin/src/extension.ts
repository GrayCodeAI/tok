import * as vscode from 'vscode';
import { TokenAnalyzer } from './analyzer';
import { StatsProvider } from './views/statsProvider';
import { HistoryProvider } from './views/historyProvider';
import { StatusBar } from './statusBar';
import { TokManAPI } from './api';

let analyzer: TokenAnalyzer;
let api: TokManAPI;
let statusBar: StatusBar;
let previewEnabled = true;

export async function activate(context: vscode.ExtensionContext) {
  console.log('TokMan VSCode extension activated');

  // Initialize components
  const config = vscode.workspace.getConfiguration('tokman');
  api = new TokManAPI(config.get('apiEndpoint') || 'http://localhost:8083');
  analyzer = new TokenAnalyzer(api, config.get('tokenCounterModel') || 'claude-3-5-sonnet');
  statusBar = new StatusBar();

  // Register commands
  registerCommands(context, analyzer);

  // Register views
  const statsProvider = new StatsProvider();
  vscode.window.registerTreeDataProvider('tokman.statsView', statsProvider);

  const historyProvider = new HistoryProvider();
  vscode.window.registerTreeDataProvider('tokman.historyView', historyProvider);

  // Register hover provider
  context.subscriptions.push(
    vscode.languages.registerHoverProvider(
      { scheme: 'file' },
      new TokenHoverProvider(analyzer, config)
    )
  );

  // Register decoration provider
  const decorationProvider = new DecorationProvider(analyzer, config);
  const activeEditor = vscode.window.activeTextEditor;
  if (activeEditor) {
    decorationProvider.updateDecorations(activeEditor);
  }

  vscode.window.onDidChangeActiveTextEditor(
    (editor) => {
      if (editor && config.get('showInlineDecorations')) {
        decorationProvider.updateDecorations(editor);
      }
    },
    null,
    context.subscriptions
  );

  vscode.workspace.onDidChangeTextDocument(
    (e) => {
      const editor = vscode.window.activeTextEditor;
      if (editor && e.document === editor.document && config.get('showInlineDecorations')) {
        decorationProvider.updateDecorations(editor);
      }
    },
    null,
    context.subscriptions
  );
}

function registerCommands(context: vscode.ExtensionContext, analyzer: TokenAnalyzer) {
  // Analyze selection
  context.subscriptions.push(
    vscode.commands.registerCommand('tokman.analyzeSelection', async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
      }

      const selection = editor.selection;
      const text = editor.document.getText(selection);

      if (!text) {
        vscode.window.showWarningMessage('No text selected');
        return;
      }

      try {
        const result = await analyzer.analyze(text);
        statusBar.show(result);

        vscode.window.showInformationMessage(
          `📊 Tokens: ${result.originalTokens} → ${result.compressedTokens} (${result.savingsPercent}% saved)`
        );
      } catch (error) {
        vscode.window.showErrorMessage(`Analysis failed: ${error}`);
      }
    })
  );

  // Analyze file
  context.subscriptions.push(
    vscode.commands.registerCommand('tokman.analyzeFile', async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
      }

      const text = editor.document.getText();

      try {
        const result = await analyzer.analyze(text);
        statusBar.show(result);

        vscode.window.showInformationMessage(
          `📊 File: ${result.originalTokens} tokens → ${result.compressedTokens} (${result.savingsPercent}% saved)`
        );
      } catch (error) {
        vscode.window.showErrorMessage(`Analysis failed: ${error}`);
      }
    })
  );

  // Toggle preview
  context.subscriptions.push(
    vscode.commands.registerCommand('tokman.togglePreview', () => {
      previewEnabled = !previewEnabled;
      const config = vscode.workspace.getConfiguration('tokman');
      config.update('realtimePreview', previewEnabled);
      statusBar.updateStatus(previewEnabled ? '🔍 Preview On' : '🔍 Preview Off');
    })
  );

  // Send to Claude
  context.subscriptions.push(
    vscode.commands.registerCommand('tokman.sendToClaude', async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
      }

      const selection = editor.selection;
      const text = editor.document.getText(selection);

      if (!text) {
        vscode.window.showWarningMessage('No text selected');
        return;
      }

      try {
        const result = await analyzer.analyze(text);
        // Open Claude in browser
        vscode.env.openExternal(vscode.Uri.parse('https://claude.ai'));
        vscode.window.showInformationMessage(
          `📋 Ready to paste: ${result.compressedTokens} tokens (was ${result.originalTokens})`
        );
      } catch (error) {
        vscode.window.showErrorMessage(`Failed to prepare content: ${error}`);
      }
    })
  );

  // Open dashboard
  context.subscriptions.push(
    vscode.commands.registerCommand('tokman.openDashboard', async () => {
      const config = vscode.workspace.getConfiguration('tokman');
      const dashboardUrl = config.get('dashboardUrl') || 'http://localhost:3000';
      vscode.env.openExternal(vscode.Uri.parse(dashboardUrl));
    })
  );

  // Settings
  context.subscriptions.push(
    vscode.commands.registerCommand('tokman.settings', () => {
      vscode.commands.executeCommand('workbench.action.openSettings', 'tokman');
    })
  );
}

// Hover provider
class TokenHoverProvider implements vscode.HoverProvider {
  constructor(private analyzer: TokenAnalyzer, private config: vscode.WorkspaceConfiguration) {}

  async provideHover(document: vscode.TextDocument, position: vscode.Position): Promise<vscode.Hover | null> {
    if (!this.config.get('realtimePreview')) {
      return null;
    }

    const wordRange = document.getWordRangeAtPosition(position);
    if (!wordRange) {
      return null;
    }

    try {
      const line = document.lineAt(position.line).text;
      const result = await this.analyzer.analyze(line);

      const markdown = new vscode.MarkdownString();
      markdown.appendMarkdown(`**Token Analysis**\n\n`);
      markdown.appendMarkdown(`- **Original**: ${result.originalTokens} tokens\n`);
      markdown.appendMarkdown(`- **Compressed**: ${result.compressedTokens} tokens\n`);
      markdown.appendMarkdown(`- **Saved**: ${result.savingsPercent}%\n`);

      return new vscode.Hover(markdown);
    } catch (error) {
      return null;
    }
  }
}

// Decoration provider
class DecorationProvider {
  private decoration: vscode.TextEditorDecorationType;

  constructor(private analyzer: TokenAnalyzer, private config: vscode.WorkspaceConfiguration) {
    this.decoration = vscode.window.createTextEditorDecorationType({
      after: {
        margin: '0 0 0 2em',
        contentText: '📊',
        backgroundColor: 'rgba(100, 200, 100, 0.1)',
        color: 'rgba(100, 200, 100, 0.7)',
        fontStyle: 'italic',
      },
      isWholeLine: false,
    });
  }

  updateDecorations(editor: vscode.TextEditor) {
    const decorations: vscode.DecorationOptions[] = [];

    for (let i = 0; i < editor.document.lineCount; i += 10) {
      decorations.push({
        range: new vscode.Range(i, 0, i, 0),
        renderOptions: {
          after: {
            contentText: '📊',
          },
        },
      });
    }

    editor.setDecorations(this.decoration, decorations);
  }
}

export function deactivate() {
  console.log('TokMan VSCode extension deactivated');
  statusBar.dispose();
}
