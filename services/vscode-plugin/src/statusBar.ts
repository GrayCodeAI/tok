import * as vscode from 'vscode';
import { AnalysisResult } from './analyzer';

export class StatusBar {
  private statusBarItem: vscode.StatusBarItem;

  constructor() {
    this.statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    this.statusBarItem.command = 'tokman.analyzeSelection';
    this.statusBarItem.tooltip = 'Click to analyze selection | Ctrl+Shift+T';
    this.statusBarItem.show();
  }

  show(result: AnalysisResult) {
    const icon = result.savingsPercent >= 50 ? '🟢' : result.savingsPercent >= 30 ? '🟡' : '🔴';
    this.statusBarItem.text = `${icon} ${result.originalTokens}→${result.compressedTokens} (${result.savingsPercent}%)`;
  }

  updateStatus(status: string) {
    this.statusBarItem.text = status;
  }

  dispose() {
    this.statusBarItem.dispose();
  }
}
