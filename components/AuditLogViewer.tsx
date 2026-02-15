
import React from 'react';
import { AuditLog, LogStatus } from '../types';

interface AuditLogViewerProps {
  logs: AuditLog[];
}

export const AuditLogViewer: React.FC<AuditLogViewerProps> = ({ logs }) => {
  const getStatusColor = (status: LogStatus) => {
    switch (status) {
      case LogStatus.SUCCESS: return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
      case LogStatus.BLOCKED: return 'bg-red-500/10 text-red-400 border-red-500/20';
      case LogStatus.CACHED: return 'bg-blue-500/10 text-blue-400 border-blue-500/20';
      case LogStatus.ERROR: return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
    }
  };

  return (
    <div className="overflow-x-auto rounded-xl border border-slate-700 bg-slate-800/50">
      <table className="w-full text-left border-collapse">
        <thead className="bg-slate-800 text-slate-400 text-sm uppercase font-semibold">
          <tr>
            <th className="px-6 py-4">Timestamp</th>
            <th className="px-6 py-4">Client</th>
            <th className="px-6 py-4">Request</th>
            <th className="px-6 py-4">Status</th>
            <th className="px-6 py-4 text-right">Latency</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-700 text-sm">
          {logs.map((log) => (
            <tr key={log.id} className="hover:bg-slate-700/30 transition-colors">
              <td className="px-6 py-4 text-slate-400 font-mono whitespace-nowrap">
                {new Date(log.timestamp).toLocaleTimeString()}
              </td>
              <td className="px-6 py-4">
                <div className="font-medium text-slate-200">{log.tokenName}</div>
                <div className="text-xs text-slate-500">{log.clientIp}</div>
              </td>
              <td className="px-6 py-4">
                <div className="flex items-center gap-2">
                  <span className="font-bold text-blue-400 text-xs">{log.method}</span>
                  <span className="text-slate-300 font-mono text-xs truncate max-w-[200px]">{log.path}</span>
                </div>
              </td>
              <td className="px-6 py-4">
                <span className={`px-2 py-1 rounded-full text-[10px] font-bold border ${getStatusColor(log.status)}`}>
                  {log.status}
                </span>
              </td>
              <td className="px-6 py-4 text-right font-mono text-slate-400">
                {log.latency}ms
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
