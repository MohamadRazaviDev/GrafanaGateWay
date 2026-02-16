
import React from 'react';
import { ApiToken, AuditLog, Dashboard, LogStatus } from './types';

export const MOCK_DASHBOARDS: Dashboard[] = [
  { id: '1', uid: 'prod-metrics', title: 'Production Overview', folderTitle: 'Core Infra' },
  { id: '2', uid: 'app-performance', title: 'User API Latency', folderTitle: 'App Ops' },
  { id: '3', uid: 'db-health', title: 'PostgreSQL Health', folderTitle: 'Database' },
  { id: '4', uid: 'cost-explorer', title: 'Cloud Spend', folderTitle: 'FinOps' },
];

export const INITIAL_TOKENS: ApiToken[] = [
  {
    id: 'tk-001',
    name: 'SRE Team Alpha',
    token: 'gg_live_7x92k...',
    team: 'Platform',
    allowedDashboards: ['prod-metrics', 'db-health'],
    rateLimit: 1000,
    createdAt: '2024-03-01T10:00:00Z',
    isActive: true,
  },
  {
    id: 'tk-002',
    name: 'Frontend Monitoring',
    token: 'gg_live_3v81m...',
    team: 'Product',
    allowedDashboards: ['app-performance'],
    rateLimit: 500,
    createdAt: '2024-03-05T14:30:00Z',
    isActive: true,
  }
];

export const generateMockLogs = (tokens: ApiToken[]): AuditLog[] => {
  const paths = ['/api/dashboards/uid/', '/api/datasources/proxy/', '/api/search', '/api/annotations'];
  const methods = ['GET', 'POST'];
  
  return Array.from({ length: 20 }).map((_, i) => {
    const token = tokens[Math.floor(Math.random() * tokens.length)];
    const statusValues = Object.values(LogStatus);
    const status = Math.random() > 0.8 ? statusValues[Math.floor(Math.random() * statusValues.length)] : LogStatus.SUCCESS;
    
    return {
      id: `log-${i}`,
      timestamp: new Date(Date.now() - i * 1000 * 60 * 5).toISOString(),
      tokenId: token.id,
      tokenName: token.name,
      path: paths[Math.floor(Math.random() * paths.length)] + (status === LogStatus.SUCCESS ? 'target-uid' : ''),
      method: methods[Math.floor(Math.random() * methods.length)],
      status: status as LogStatus,
      latency: Math.floor(Math.random() * 200) + 20,
      clientIp: `192.168.1.${Math.floor(Math.random() * 255)}`,
    };
  });
};
