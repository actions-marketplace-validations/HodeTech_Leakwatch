/**
 * Scanner module — executes the leakwatch binary and parses results.
 */

import * as vscode from "vscode";
import { execFile } from "child_process";
import { Finding } from "./types";

/** Result of a scan operation. */
export interface ScanResult {
  findings: Finding[];
  error?: string;
}

/**
 * Runs leakwatch scan on the given path and returns parsed findings.
 * Uses JSON output format for reliable parsing.
 */
export function scan(
  targetPath: string,
  token: vscode.CancellationToken
): Promise<ScanResult> {
  return new Promise((resolve) => {
    const config = vscode.workspace.getConfiguration("leakwatch");
    const executable = config.get<string>("executablePath", "leakwatch");
    const minSeverity = config.get<string>("minSeverity", "low");
    const customRulesPath = config.get<string>("customRulesPath", "");

    const args = [
      "scan",
      "fs",
      targetPath,
      "--format",
      "json",
      "--min-severity",
      minSeverity,
      "--no-verify",
    ];

    if (customRulesPath) {
      args.push("--config", customRulesPath);
    }

    const process = execFile(
      executable,
      args,
      { maxBuffer: 10 * 1024 * 1024, timeout: 60_000 },
      (error, stdout, stderr) => {
        if (token.isCancellationRequested) {
          resolve({ findings: [] });
          return;
        }

        // Exit code 1 means findings were detected (not an error)
        if (error && error.code !== 1) {
          const message = stderr.trim() || error.message;
          resolve({ findings: [], error: message });
          return;
        }

        const findings = parseFindings(stdout);
        resolve({ findings });
      }
    );

    token.onCancellationRequested(() => {
      process.kill();
    });
  });
}

/**
 * Parses leakwatch JSON output into Finding objects.
 * Returns an empty array if the output cannot be parsed.
 */
function parseFindings(output: string): Finding[] {
  const trimmed = output.trim();
  if (!trimmed || trimmed === "null") {
    return [];
  }

  try {
    const parsed = JSON.parse(trimmed);
    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.map(
      (f: Record<string, unknown>): Finding => ({
        id: String(f.id ?? ""),
        detectorId: String(f.detector_id ?? ""),
        description: String(f.description ?? ""),
        severity: normalizeSeverity(f.severity),
        redacted: String(f.redacted ?? ""),
        verificationStatus: String(
          f.verification_status ?? "unverified"
        ) as Finding["verificationStatus"],
        sourceMetadata: {
          filePath: String(
            (f.source_metadata as Record<string, unknown>)?.file_path ?? ""
          ),
          lineNumber: Number(
            (f.source_metadata as Record<string, unknown>)?.line_number ?? 0
          ),
          sourceType: String(
            (f.source_metadata as Record<string, unknown>)?.source_type ?? ""
          ),
        },
        extraData: (f.extra_data as Record<string, string>) ?? undefined,
      })
    );
  } catch {
    return [];
  }
}

function normalizeSeverity(value: unknown): Finding["severity"] {
  const s = String(value).toLowerCase();
  if (s === "low" || s === "medium" || s === "high" || s === "critical") {
    return s;
  }
  return "medium";
}
