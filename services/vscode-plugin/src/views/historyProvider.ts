import * as vscode from 'vscode';

export interface HistoryEntry {
  id: string;
  timestamp: Date;
  originalTokens: number;
  compressedTokens: number;
  savingsPercent: number;
  preview: string;
}

export class HistoryProvider implements vscode.TreeDataProvider<HistoryItem> {
  private _onDidChangeTreeData: vscode.EventEmitter<HistoryItem | undefined | void> = new vscode.EventEmitter<HistoryItem | undefined | void>();
  readonly onDidChangeTreeData: vscode.Event<HistoryItem | undefined | void> = this._onDidChangeTreeData.event;

  private history: HistoryEntry[] = [];

  refresh(): void {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: HistoryItem): vscode.TreeItem {
    return element;
  }

  getChildren(element?: HistoryItem): Thenable<HistoryItem[]> {
    if (element === undefined) {
      return Promise.resolve(
        this.history.slice(0, 20).map((entry) => {
          const time = entry.timestamp.toLocaleTimeString();
          const label = `${time} - ${entry.originalTokens}→${entry.compressedTokens} (${entry.savingsPercent}%)`;
          return new HistoryItem(
            label,
            entry.preview.substring(0, 50),
            vscode.TreeItemCollapsibleState.None,
            entry.id
          );
        })
      );
    }

    return Promise.resolve([]);
  }

  addEntry(entry: HistoryEntry) {
    this.history.unshift(entry);
    // Keep only last 100 entries
    if (this.history.length > 100) {
      this.history.pop();
    }
    this.refresh();
  }

  clearHistory() {
    this.history = [];
    this.refresh();
  }

  getEntries(): HistoryEntry[] {
    return this.history;
  }
}

class HistoryItem extends vscode.TreeItem {
  constructor(
    public readonly label: string,
    public readonly description: string,
    public readonly collapsibleState: vscode.TreeItemCollapsibleState,
    public readonly id: string,
  ) {
    super(label, collapsibleState);
    this.description = description;
    this.contextValue = 'history';
  }
}
