
export enum LogStatus {
  SUCCESS = 'SUCCESS',
  BLOCKED = 'BLOCKED',
  CACHED = 'CACHED',
  ERROR = 'ERROR'
}

export interface ApiToken {
  id: string;
  name: string;
  token: string;
  team: string;
  allowedDashboards: string[];
  rateLimit: number; // requests per minute
  createdAt: string;
  isActive: boolean;
}

export interface AuditLog {
  id: string;
  timestamp: string;
  tokenId: string;
  tokenName: string;
  path: string;
  method: string;
  status: LogStatus;
  latency: number; // ms
  clientIp: string;
}

export interface GatewayStats {
  totalRequests: number;
  blockedRequests: number;
  cacheHitRate: number;
  activeTokens: number;
}

export interface Dashboard {
  id: string;
  uid: string;
  title: string;
  folderTitle: string;
}
