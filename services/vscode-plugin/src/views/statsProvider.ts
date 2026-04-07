import * as vscode from 'vscode';

export class StatsProvider implements vscode.TreeDataProvider<StatsItem> {
  private _onDidChangeTreeData: vscode.EventEmitter<StatsItem | undefined | void> = new vscode.EventEmitter<StatsItem | undefined | void>();
  readonly onDidChangeTreeData: vscode.Event<StatsItem | undefined | void> = this._onDidChangeTreeData.event;

  private stats = {
    sessionsAnalyzed: 0,
    totalTokensAnalyzed: 0,
    totalTokensSaved: 0,
    averageSavingsPercent: 0,
  };

  refresh(): void {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: StatsItem): vscode.TreeItem {
    return element;
  }

  getChildren(element?: StatsItem): Thenable<StatsItem[]> {
    if (element === undefined) {
      return Promise.resolve([
        new StatsItem('📊 Sessions Analyzed', `${this.stats.sessionsAnalyzed}`, vscode.TreeItemCollapsibleState.None),
        new StatsItem('📈 Total Tokens Analyzed', `${this.stats.totalTokensAnalyzed}`, vscode.TreeItemCollapsibleState.None),
        new StatsItem('✅ Total Tokens Saved', `${this.stats.totalTokensSaved}`, vscode.TreeItemCollapsibleState.None),
        new StatsItem('🎯 Average Savings', `${this.stats.averageSavingsPercent}%`, vscode.TreeItemCollapsibleState.None),
      ]);
    }

    return Promise.resolve([]);
  }

  updateStats(updates: Partial<typeof this.stats>) {
    this.stats = { ...this.stats, ...updates };
    this.refresh();
  }
}

class StatsItem extends vscode.TreeItem {
  constructor(
    public readonly label: string,
    public readonly value: string,
    public readonly collapsibleState: vscode.TreeItemCollapsibleState,
  ) {
    super(label, collapsibleState);
    this.description = value;
    this.contextValue = 'stat';
  }
}
