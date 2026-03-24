/**
 * Diagnostics module — converts Leakwatch findings to VS Code diagnostics.
 */

import * as vscode from "vscode";
import { Finding, Severity } from "./types";

/** Maps Leakwatch severity to VS Code diagnostic severity. */
function toVSCodeSeverity(severity: Severity): vscode.DiagnosticSeverity {
  switch (severity) {
    case "critical":
      return vscode.DiagnosticSeverity.Error;
    case "high":
      return vscode.DiagnosticSeverity.Error;
    case "medium":
      return vscode.DiagnosticSeverity.Warning;
    case "low":
      return vscode.DiagnosticSeverity.Information;
  }
}

/** Creates a VS Code diagnostic from a Leakwatch finding. */
function toDiagnostic(finding: Finding): vscode.Diagnostic {
  const line = Math.max(0, finding.sourceMetadata.lineNumber - 1);
  const range = new vscode.Range(line, 0, line, Number.MAX_SAFE_INTEGER);

  const severity = toVSCodeSeverity(finding.severity);
  const message = `[${finding.severity.toUpperCase()}] ${finding.description}: ${finding.redacted}`;

  const diagnostic = new vscode.Diagnostic(range, message, severity);
  diagnostic.source = "Leakwatch";
  diagnostic.code = finding.detectorId;

  return diagnostic;
}

/**
 * Updates the diagnostic collection with findings grouped by file.
 * Clears previous diagnostics before applying new ones.
 */
export function updateDiagnostics(
  collection: vscode.DiagnosticCollection,
  findings: Finding[],
  workspacePath: string
): void {
  collection.clear();

  // Group findings by file path
  const grouped = new Map<string, vscode.Diagnostic[]>();

  for (const finding of findings) {
    const filePath = finding.sourceMetadata.filePath;
    if (!filePath) {
      continue;
    }

    // Resolve to absolute path if relative
    const absolutePath = filePath.startsWith("/")
      ? filePath
      : `${workspacePath}/${filePath}`;

    const uri = vscode.Uri.file(absolutePath);
    const key = uri.toString();

    if (!grouped.has(key)) {
      grouped.set(key, []);
    }

    grouped.get(key)!.push(toDiagnostic(finding));
  }

  for (const [uriString, diagnostics] of grouped) {
    collection.set(vscode.Uri.parse(uriString), diagnostics);
  }
}
