/**
 * Leakwatch VS Code Extension
 *
 * Integrates the Leakwatch secret scanner with VS Code, providing:
 * - Real-time diagnostics (Problems panel) for detected secrets
 * - Scan on save (configurable)
 * - Manual workspace and file scanning via commands
 * - Status bar indicator showing scan state and finding count
 */

import * as vscode from "vscode";
import * as path from "path";
import { scan } from "./scanner";
import { updateDiagnostics } from "./diagnostics";
import * as statusbar from "./statusbar";

let diagnosticCollection: vscode.DiagnosticCollection;

/** Debounce timer for scan-on-save. */
let saveTimer: ReturnType<typeof setTimeout> | undefined;
const SAVE_DEBOUNCE_MS = 1000;

export function activate(context: vscode.ExtensionContext): void {
  diagnosticCollection =
    vscode.languages.createDiagnosticCollection("leakwatch");
  context.subscriptions.push(diagnosticCollection);

  const bar = statusbar.createStatusBar();
  context.subscriptions.push(bar);

  // Register commands
  context.subscriptions.push(
    vscode.commands.registerCommand("leakwatch.scanWorkspace", () =>
      scanWorkspace()
    ),
    vscode.commands.registerCommand("leakwatch.scanCurrentFile", () =>
      scanCurrentFile()
    ),
    vscode.commands.registerCommand("leakwatch.clearDiagnostics", () => {
      diagnosticCollection.clear();
      statusbar.setIdle();
    })
  );

  // Scan on save
  context.subscriptions.push(
    vscode.workspace.onDidSaveTextDocument((document) => {
      const config = vscode.workspace.getConfiguration("leakwatch");
      if (!config.get<boolean>("scanOnSave", true)) {
        return;
      }

      // Debounce rapid saves
      if (saveTimer) {
        clearTimeout(saveTimer);
      }
      saveTimer = setTimeout(() => {
        scanFile(document.uri.fsPath);
      }, SAVE_DEBOUNCE_MS);
    })
  );
}

export function deactivate(): void {
  if (saveTimer) {
    clearTimeout(saveTimer);
  }
}

/**
 * Scans the entire workspace folder.
 */
async function scanWorkspace(): Promise<void> {
  const workspaceFolders = vscode.workspace.workspaceFolders;
  if (!workspaceFolders || workspaceFolders.length === 0) {
    vscode.window.showWarningMessage("Leakwatch: No workspace folder open.");
    return;
  }

  const workspacePath = workspaceFolders[0].uri.fsPath;

  statusbar.setScanning();

  const tokenSource = new vscode.CancellationTokenSource();
  const result = await scan(workspacePath, tokenSource.token);
  tokenSource.dispose();

  if (result.error) {
    statusbar.setError(result.error);
    vscode.window.showErrorMessage(`Leakwatch: ${result.error}`);
    return;
  }

  updateDiagnostics(diagnosticCollection, result.findings, workspacePath);
  statusbar.setResults(result.findings.length);

  if (result.findings.length > 0) {
    vscode.window.showWarningMessage(
      `Leakwatch: Found ${result.findings.length} secret${result.findings.length === 1 ? "" : "s"}. Check the Problems panel.`
    );
  }
}

/**
 * Scans the currently active file.
 */
async function scanCurrentFile(): Promise<void> {
  const editor = vscode.window.activeTextEditor;
  if (!editor) {
    vscode.window.showWarningMessage("Leakwatch: No active file.");
    return;
  }

  await scanFile(editor.document.uri.fsPath);
}

/**
 * Scans a single file by creating a temporary directory context.
 */
async function scanFile(filePath: string): Promise<void> {
  const workspaceFolders = vscode.workspace.workspaceFolders;
  const workspacePath = workspaceFolders?.[0]?.uri.fsPath ?? path.dirname(filePath);
  const dir = path.dirname(filePath);

  statusbar.setScanning();

  const tokenSource = new vscode.CancellationTokenSource();
  const result = await scan(dir, tokenSource.token);
  tokenSource.dispose();

  if (result.error) {
    statusbar.setError(result.error);
    return;
  }

  // Filter to only findings from this specific file
  const fileFindings = result.findings.filter((f) => {
    const findingPath = f.sourceMetadata.filePath;
    return (
      findingPath === filePath ||
      findingPath === path.basename(filePath) ||
      filePath.endsWith(findingPath)
    );
  });

  // Clear diagnostics only for this file, then set new ones
  const fileUri = vscode.Uri.file(filePath);
  diagnosticCollection.delete(fileUri);

  updateDiagnostics(diagnosticCollection, fileFindings, workspacePath);
  statusbar.setResults(fileFindings.length);
}
