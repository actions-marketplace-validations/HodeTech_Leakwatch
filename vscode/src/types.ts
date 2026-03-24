/**
 * Types shared across the Leakwatch VS Code extension.
 */

/** Severity levels matching Leakwatch CLI output. */
export type Severity = "low" | "medium" | "high" | "critical";

/** Verification status from Leakwatch verifier engine. */
export type VerificationStatus =
  | "unverified"
  | "verified_active"
  | "verified_inactive"
  | "verify_error";

/** Source metadata describing where a finding originated. */
export interface SourceMetadata {
  filePath: string;
  lineNumber: number;
  sourceType: string;
  commitHash?: string;
  commitAuthor?: string;
  commitDate?: string;
}

/** A single finding from Leakwatch JSON output. */
export interface Finding {
  id: string;
  detectorId: string;
  description: string;
  severity: Severity;
  redacted: string;
  verificationStatus: VerificationStatus;
  sourceMetadata: SourceMetadata;
  extraData?: Record<string, string>;
}
