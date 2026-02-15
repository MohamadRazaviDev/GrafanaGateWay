
import React, { useState, useEffect, useMemo } from 'react';
import { StatCard } from './components/StatCard';
import { AuditLogViewer } from './components/AuditLogViewer';
import { TokenManager } from './components/TokenManager';
import { ApiToken, AuditLog, GatewayStats, LogStatus } from './types';
import { INITIAL_TOKENS, generateMockLogs } from './constants';
import { analyzeSecurityLogs } from './services/geminiService';

const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'overview' | 'tokens' | 'logs' | 'ai'>('overview');
  const [tokens, setTokens] = useState<ApiToken[]>(INITIAL_TOKENS);
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [aiAnalysis, setAiAnalysis] = useState<any>(null);
  const [isAnalyzing, setIsAnalyzing] = useState(false);

  useEffect(() => {
    setLogs(generateMockLogs(tokens));
  }, []);

  const stats: GatewayStats = useMemo(() => {
    return {
      totalRequests: logs.length * 142, // Scaled for effect
      blockedRequests: logs.filter(l => l.status === LogStatus.BLOCKED).length * 12,
      cacheHitRate: 42.5,
      activeTokens: tokens.filter(t => t.isActive).length,
    };
  }, [logs, tokens]);

  const handleAddToken = (tokenData: any) => {
    const newToken: ApiToken = {
      ...tokenData,
      id: `tk-${Math.random().toString(36).substring(7)}`,
      createdAt: new Date().toISOString(),
      isActive: true,
    };
    setTokens(prev => [...prev, newToken]);
  };

  const handleDeleteToken = (id: string) => {
    setTokens(prev => prev.filter(t => t.id !== id));
  };

  const handleRunAiAnalysis = async () => {
    setIsAnalyzing(true);
    try {
      const result = await analyzeSecurityLogs(logs);
      setAiAnalysis(result);
      setActiveTab('ai');
    } catch (error) {
      console.error("AI Analysis failed", error);
    } finally {
      setIsAnalyzing(false);
    }
  };

  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="w-64 border-r border-slate-800 bg-slate-900/50 backdrop-blur-md sticky top-0 h-screen flex flex-col">
        <div className="p-6">
          <div className="flex items-center gap-3 mb-8">
            <div className="w-10 h-10 bg-orange-500 rounded-lg flex items-center justify-center shadow-lg shadow-orange-500/20">
              <svg className="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
            </div>
            <div>
              <h1 className="font-bold text-lg leading-tight">Grafana</h1>
              <span className="text-orange-400 text-xs font-bold tracking-widest uppercase">Gateway</span>
            </div>
          </div>

          <nav className="space-y-1">
            <button 
              onClick={() => setActiveTab('overview')}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all ${activeTab === 'overview' ? 'bg-slate-800 text-white shadow-inner border border-slate-700' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'}`}
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>
              Overview
            </button>
            <button 
              onClick={() => setActiveTab('tokens')}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all ${activeTab === 'tokens' ? 'bg-slate-800 text-white shadow-inner border border-slate-700' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'}`}
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" /></svg>
              Tokens & RBAC
            </button>
            <button 
              onClick={() => setActiveTab('logs')}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all ${activeTab === 'logs' ? 'bg-slate-800 text-white shadow-inner border border-slate-700' : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'}`}
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
              Audit Logs
            </button>
          </nav>

          <div className="mt-8 pt-8 border-t border-slate-800">
             <button 
              onClick={handleRunAiAnalysis}
              disabled={isAnalyzing}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-bold transition-all bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 text-white shadow-lg shadow-purple-500/20 disabled:opacity-50`}
            >
              {isAnalyzing ? (
                <svg className="animate-spin h-5 w-5 mr-3 text-white" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              ) : (
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
              )}
              {isAnalyzing ? "Analyzing..." : "Gemini AI Scan"}
            </button>
          </div>
        </div>

        <div className="mt-auto p-6 bg-slate-900/50">
          <div className="flex items-center gap-3 text-sm text-slate-400">
            <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            Proxy Status: Healthy
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 p-8 overflow-y-auto">
        <header className="flex justify-between items-center mb-8">
          <div>
            <h2 className="text-3xl font-bold text-white">
              {activeTab === 'overview' && 'Dashboard Overview'}
              {activeTab === 'tokens' && 'Token Management'}
              {activeTab === 'logs' && 'Real-time Traffic'}
              {activeTab === 'ai' && 'AI Security Report'}
            </h2>
            <p className="text-slate-400 mt-1">
              {activeTab === 'overview' && 'System-wide proxy metrics and status.'}
              {activeTab === 'tokens' && 'Manage team access and rate limit policies.'}
              {activeTab === 'logs' && 'Auditing and troubleshooting the gateway proxy.'}
              {activeTab === 'ai' && 'Gemini-powered insights into your proxy traffic.'}
            </p>
          </div>
          <div className="flex items-center gap-4">
            <div className="bg-slate-800 border border-slate-700 px-4 py-2 rounded-lg text-sm font-medium text-slate-300">
              Region: us-east-1
            </div>
          </div>
        </header>

        {activeTab === 'overview' && (
          <div className="space-y-8 animate-fadeIn">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <StatCard label="Total Requests" value={stats.totalRequests.toLocaleString()} subValue="+12% from last week" icon={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>} />
              <StatCard label="Cache Hit Rate" value={`${stats.cacheHitRate}%`} icon={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M4 7v10c0 2 1.5 3 3.5 3h9c2 0 3.5-1 3.5-3V7c0-2-1.5-3-3.5-3h-9C5.5 4 4 5 4 7z" /></svg>} colorClass="text-emerald-400" />
              <StatCard label="Blocked Reqs" value={stats.blockedRequests} icon={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>} colorClass="text-rose-400" />
              <StatCard label="Active Tokens" value={stats.activeTokens} icon={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A10.003 10.003 0 0012 3m0 18a10.003 10.003 0 01-12-10V7a2 2 0 012-2h2m2 4h6a2 2 0 012 2v8a2 2 0 01-2 2H13" /></svg>} colorClass="text-indigo-400" />
            </div>

            <div className="bg-slate-800/50 border border-slate-700 rounded-2xl p-6">
              <div className="flex justify-between items-center mb-6">
                <h3 className="font-bold text-lg text-white">Recent Security Activity</h3>
                <button onClick={() => setActiveTab('logs')} className="text-blue-400 text-sm font-medium hover:underline">View All Logs</button>
              </div>
              <AuditLogViewer logs={logs.slice(0, 8)} />
            </div>
          </div>
        )}

        {activeTab === 'tokens' && (
          <TokenManager tokens={tokens} onAddToken={handleAddToken} onDeleteToken={handleDeleteToken} />
        )}

        {activeTab === 'logs' && (
          <div className="space-y-6">
             <div className="flex gap-4 mb-4">
                <input type="text" placeholder="Search paths or tokens..." className="flex-1 bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white outline-none focus:border-blue-500" />
                <select className="bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white outline-none focus:border-blue-500">
                  <option>All Statuses</option>
                  <option>SUCCESS</option>
                  <option>BLOCKED</option>
                </select>
             </div>
             <AuditLogViewer logs={logs} />
          </div>
        )}

        {activeTab === 'ai' && aiAnalysis && (
          <div className="space-y-8 animate-fadeIn">
            <div className={`p-6 rounded-2xl border ${aiAnalysis.threatLevel === 'HIGH' ? 'bg-rose-500/10 border-rose-500/20' : 'bg-emerald-500/10 border-emerald-500/20'}`}>
              <div className="flex items-center gap-4 mb-4">
                <div className={`p-3 rounded-xl ${aiAnalysis.threatLevel === 'HIGH' ? 'bg-rose-500 text-white' : 'bg-emerald-500 text-white'}`}>
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.952 11.952 0 01-8.618 3.04l-.53.057C2.11 6.17 1 7.36 1 8.8v3.31c0 6.637 4.135 12.5 10.424 14.841a1.13 1.13 0 00.752 0C18.465 24.61 22.56 18.747 22.56 12.11V8.8c0-1.44-1.11-2.63-2.512-2.763l-.43-.037z" /></svg>
                </div>
                <div>
                  <h3 className="text-xl font-bold text-white">Threat Level: {aiAnalysis.threatLevel}</h3>
                  <p className="text-slate-400">Analysis completed based on the last 100 requests.</p>
                </div>
              </div>
              <p className="text-slate-200 leading-relaxed mb-6">{aiAnalysis.summary}</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <div className="bg-slate-800 border border-slate-700 p-6 rounded-2xl">
                <h4 className="font-bold text-white mb-4 flex items-center gap-2">
                  <svg className="w-5 h-5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                  Detected Anomalies
                </h4>
                <ul className="space-y-3">
                  {aiAnalysis.anomalies.map((item: string, i: number) => (
                    <li key={i} className="flex gap-3 text-sm text-slate-300">
                      <span className="text-slate-500 font-mono">{i+1}.</span>
                      {item}
                    </li>
                  ))}
                </ul>
              </div>
              <div className="bg-slate-800 border border-slate-700 p-6 rounded-2xl">
                <h4 className="font-bold text-white mb-4 flex items-center gap-2">
                  <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
                  Recommended Actions
                </h4>
                <ul className="space-y-3">
                  {aiAnalysis.recommendations.map((item: string, i: number) => (
                    <li key={i} className="flex gap-3 text-sm text-slate-300">
                      <span className="bg-blue-500/20 text-blue-400 w-5 h-5 flex items-center justify-center rounded-full text-[10px] flex-shrink-0">✓</span>
                      {item}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
};

export default App;
