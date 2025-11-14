export type Phase = "UNKNOWN" | "READY" | "PENDING" | "PROCESSING" | "FAILED";

export interface ModelSummary {
  name: string;
  namespace: string;
  sourceRef: string;
  tags?: string[];
  versionsReady: number;
  versionsTotal: number;
  lastSyncTime: string;
  status: Phase;
  reconcileMessage?: string;
}

export interface ModelVersionView {
  name: string;
  repo: string;
  revision?: string;
  precision?: string;
  desiredState: string;
  shareEnabled: boolean;
  namespacePolicy?: string;
  datasetPhase: Phase;
  pvcName?: string;
  observedHash?: string;
}

export interface ModelDetail {
  summary: ModelSummary;
  description?: string;
  logoURL?: string;
  versions: ModelVersionView[];
  shareTargets?: string[];
  conditionsJson?: string;
}

export interface ModelSourceSummary {
  name: string;
  namespace: string;
  type: string;
  secretRef?: string;
  credentialsReady: boolean;
  credentialsStatus?: string;
  referencedModels?: string[];
  lastChecked: string;
}

export interface NamespaceInfo { name: string }

export interface ErrorBanner {
  namespace: string;
  message: string;
  reason: string;
  retryAt: string;
}

