export type ScanStatus = "QUEUED" | "RUNNING" | "SUCCEEDED" | "FAILED" | "CANCELED";
export type ScanType = "SCA" | "SAST";

export interface User {
  id: string;
  username: string;
  displayName: string;
  role: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  ownerUsername: string;
  createdAt: string;
}

export interface Scan {
  id: string;
  projectId: string;
  scanType: ScanType;
  status: ScanStatus;
  sourceType: "UPLOAD" | "GIT";
  sourceRef: string;
  sourceRevision: string;
  sourceMode: string;
  engineVersion: string;
  failureCode: string;
  failureMessage: string;
  componentCount: number;
  findingCount: number;
  sourceExpiresAt: string;
  sourceDeleted: boolean;
  resultExpiresAt: string;
  createdAt: string;
  startedAt: string;
  finishedAt: string;
}

export interface Component {
  id: string;
  ecosystem: string;
  groupId: string;
  artifactId: string;
  version: string;
  purl: string;
  direct: boolean;
  dependencyPath: string;
  scope: string;
}

export interface Finding {
  id: string;
  findingType: ScanType;
  vulnerabilityId: string;
  ruleId: string;
  aliases: string;
  title: string;
  severity: string;
  confidence: string;
  category: string;
  cwe: string;
  filePath: string;
  startLine: number;
  endLine: number;
  codeSnippet: string;
  recommendation: string;
  packageName: string;
  affectedVersion: string;
  affectedRange: string;
  fixedVersions: string;
  dependencyPath: string;
  createdAt: string;
}
