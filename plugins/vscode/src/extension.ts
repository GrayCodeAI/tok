import * as vscode from 'vscode';
import { spawn, exec } from 'child_process';
import * as path from 'path';
import * as fs from 'fs';

let tokmanStatusBar: vscode.StatusBarItem;
let tokmanEnabled = true;
let tokmanPath: string | null = null;

export function activate(context: vscode.ExtensionContext) {
    console.log('TokMan extension activating...');
    
    findTokmanBinary();
    setupStatusBar();
    setupCommands(context);
    setupTerminalIntegration();
    setupConfiguration();
    
    console.log('TokMan extension activated successfully');
}

function findTokmanBinary() {
    const possiblePaths = [
        path.join(process.env.HOME || '', 'go/bin/tokman'),
        '/usr/local/bin/tokman',
        '/usr/bin/tokman',
        path.join(__dirname, '../../../tokman')
    ];
    
    for (const p of possiblePaths) {
        if (fs.existsSync(p)) {
            tokmanPath = p;
            console.log('Found TokMan at:', p);
            break;
        }
    }
    
    if (!tokmanPath) {
        vscode.window.showWarningMessage('TokMan binary not found. Run "go install github.com/GrayCodeAI/tokman@latest"');
    }
}

function setupStatusBar() {
    tokmanStatusBar = vscode.window.createStatusBarItem(
        vscode.StatusBarAlignment.Right,
        100
    );
    tokmanStatusBar.text = '$(cloud) TokMan';
    tokmanStatusBar.tooltip = 'TokMan: Token Optimizer';
    tokmanStatusBar.command = 'tokman.status';
    tokmanStatusBar.show();
}

function setupCommands(context: vscode.ExtensionContext) {
    const commands = [
        { cmd: 'tokman.init', handler: () => runTokmanCommand(['init']) },
        { cmd: 'tokman.enable', handler: () => setEnabled(true) },
        { cmd: 'tokman.disable', handler: () => setEnabled(false) },
        { cmd: 'tokman.status', handler: () => showStatus() },
        { cmd: 'tokman.stats', handler: () => showStats() },
        { cmd: 'tokman.dashboard', handler: () => openDashboard() },
        { cmd: 'tokman.config', handler: () => openConfig() },
        { cmd: 'tokman.compressTest', handler: () => testCompression() },
        { cmd: 'tokman.setBudget', handler: () => setBudget() },
        { cmd: 'tokman.setProfile', handler: () => setProfile() },
    ];
    
    commands.forEach(({ cmd, handler }) => {
        const disposable = vscode.commands.registerCommand(cmd, handler);
        context.subscriptions.push(disposable);
    });
}

function setupTerminalIntegration() {
    const config = vscode.workspace.getConfiguration('tokman');
    if (config.get('terminalIntegration', true)) {
        vscode.window.onDidOpenTerminal(async (terminal) => {
            if (tokmanEnabled && tokmanPath) {
                await vscode.commands.executeCommand('tokman.init');
            }
        });
    }
}

function setupConfiguration() {
    vscode.workspace.onDidChangeConfiguration((e) => {
        if (e.affectsConfiguration('tokman')) {
            const config = vscode.workspace.getConfiguration('tokman');
            tokmanEnabled = config.get('enabled', true);
            updateStatusBar();
        }
    });
}

function setEnabled(enabled: boolean) {
    tokmanEnabled = enabled;
    vscode.workspace.getConfiguration('tokman').update('enabled', enabled, true);
    updateStatusBar();
    vscode.window.showInformationMessage(`TokMan ${enabled ? 'enabled' : 'disabled'}`);
}

function updateStatusBar() {
    tokmanStatusBar.text = tokmanEnabled ? '$(cloud) TokMan' : '$(cloud-off) TokMan';
}

async function showStatus() {
    if (!tokmanPath) {
        vscode.window.showWarningMessage('TokMan not found');
        return;
    }
    
    const result = await runTokmanCommand(['status']);
    vscode.window.showInformationMessage(result || 'TokMan is running');
}

async function showStats() {
    if (!tokmanPath) {
        vscode.window.showWarningMessage('TokMan not found');
        return;
    }
    
    const panel = vscode.window.createWebviewPanel(
        'tokman-stats',
        'TokMan Statistics',
        vscode.ViewColumn.One,
        {}
    );
    
    const result = await runTokmanCommand(['analytics', 'top']);
    panel.webview.html = getStatsHtml(result || 'No data');
}

function getStatsHtml(data: string): string {
    return `
    <!DOCTYPE html>
    <html>
    <head>
        <style>
            body { font-family: system-ui; padding: 20px; background: #1a1a2e; color: #eee; }
            pre { background: #16213e; padding: 15px; border-radius: 8px; overflow-x: auto; }
        </style>
    </head>
    <body>
        <h1>TokMan Statistics</h1>
        <pre>${data}</pre>
    </body>
    </html>
    `;
}

async function openDashboard() {
    const config = vscode.workspace.getConfiguration('tokman');
    const port = config.get('dashboardPort', 8080);
    
    vscode.env.openExternal(vscode.Uri.parse(`http://localhost:${port}`));
}

async function openConfig() {
    const configPath = path.join(process.env.HOME || '', '.config/tokman/config.toml');
    const doc = await vscode.workspace.openTextDocument(configPath);
    await vscode.window.showTextDocument(doc);
}

async function testCompression() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) return;
    
    const selection = editor.document.getText(editor.selection);
    if (!selection) {
        vscode.window.showInformationMessage('Select text to test compression');
        return;
    }
    
    if (!tokmanPath) {
        vscode.window.showWarningMessage('TokMan not found');
        return;
    }
    
    const result = await runTokmanCommand(['compress'], selection);
    
    const panel = vscode.window.createWebviewPanel(
        'tokman-compress-test',
        'Compression Test',
        vscode.ViewColumn.Two,
        {}
    );
    
    panel.webview.html = `
    <!DOCTYPE html>
    <html>
    <head>
        <style>
            body { font-family: monospace; padding: 20px; background: #1a1a2e; color: #eee; }
            .original, .compressed { padding: 10px; margin: 10px 0; background: #16213e; border-radius: 8px; }
            .label { color: #4ade80; font-weight: bold; }
        </style>
    </head>
    <body>
        <div class="original"><span class="label">Original:</span><br/>${selection.substring(0, 500)}</div>
        <div class="compressed"><span class="label">Compressed:</span><br/>${result.substring(0, 500)}</div>
    </body>
    </html>
    `;
}

async function setBudget() {
    const budget = await vscode.window.showInputBox({
        prompt: 'Enter token budget (0 = unlimited)',
        value: '2000'
    });
    
    if (budget) {
        await vscode.workspace.getConfiguration('tokman').update('budget', parseInt(budget), true);
        vscode.window.showInformationMessage(`Token budget set to ${budget}`);
    }
}

async function setProfile() {
    const profile = await vscode.window.showQuickPick(['fast', 'balanced', 'full'], {
        placeHolder: 'Select compression profile'
    });
    
    if (profile) {
        await vscode.workspace.getConfiguration('tokman').update('profile', profile, true);
        vscode.window.showInformationMessage(`Profile set to ${profile}`);
    }
}

function runTokmanCommand(args: string[], stdin?: string): Promise<string> {
    return new Promise((resolve, reject) => {
        if (!tokmanPath) {
            reject(new Error('TokMan not found'));
            return;
        }
        
        const cmd = spawn(tokmanPath, args);
        let stdout = '';
        let stderr = '';
        
        if (stdin) {
            cmd.stdin.write(stdin);
            cmd.stdin.end();
        }
        
        cmd.stdout.on('data', (data) => { stdout += data.toString(); });
        cmd.stderr.on('data', (data) => { stderr += data.toString(); });
        
        cmd.on('close', (code) => {
            if (code === 0) {
                resolve(stdout);
            } else {
                reject(new Error(stderr || `Exit code: ${code}`));
            }
        });
        
        cmd.on('error', reject);
    });
}

export function deactivate() {
    console.log('TokMan extension deactivated');
}