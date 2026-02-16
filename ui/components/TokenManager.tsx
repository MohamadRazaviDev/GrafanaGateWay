
import React, { useState } from 'react';
import { ApiToken, Dashboard } from '../types';
import { MOCK_DASHBOARDS } from '../constants';

interface TokenManagerProps {
  tokens: ApiToken[];
  onAddToken: (token: Omit<ApiToken, 'id' | 'createdAt' | 'isActive'>) => void;
  onDeleteToken: (id: string) => void;
}

export const TokenManager: React.FC<TokenManagerProps> = ({ tokens, onAddToken, onDeleteToken }) => {
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    team: '',
    rateLimit: 100,
    allowedDashboards: [] as string[]
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onAddToken({
      ...formData,
      token: `gg_live_${Math.random().toString(36).substring(7)}...`
    });
    setFormData({ name: '', team: '', rateLimit: 100, allowedDashboards: [] });
    setShowModal(false);
  };

  const toggleDashboard = (uid: string) => {
    setFormData(prev => ({
      ...prev,
      allowedDashboards: prev.allowedDashboards.includes(uid)
        ? prev.allowedDashboards.filter(u => u !== uid)
        : [...prev.allowedDashboards, uid]
    }));
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-bold text-white">Access Tokens & RBAC</h2>
        <button 
          onClick={() => setShowModal(true)}
          className="bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded-lg font-medium transition-colors"
        >
          Create New Token
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {tokens.map(token => (
          <div key={token.id} className="bg-slate-800 border border-slate-700 p-5 rounded-xl flex flex-col justify-between">
            <div>
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h4 className="font-bold text-lg text-white">{token.name}</h4>
                  <span className="text-xs text-blue-400 uppercase tracking-tighter bg-blue-500/10 px-2 py-0.5 rounded border border-blue-500/20">{token.team}</span>
                </div>
                <button 
                  onClick={() => onDeleteToken(token.id)}
                  className="text-slate-500 hover:text-red-400"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
              </div>
              <div className="space-y-3 mb-6">
                <div>
                  <p className="text-xs text-slate-500 mb-1">Token Identifier</p>
                  <code className="text-xs bg-slate-900 px-2 py-1 rounded block text-slate-300 font-mono">{token.token}</code>
                </div>
                <div>
                  <p className="text-xs text-slate-500 mb-1">Rate Limit</p>
                  <p className="text-sm font-medium text-slate-200">{token.rateLimit} req/min</p>
                </div>
                <div>
                  <p className="text-xs text-slate-500 mb-1">Allowed Dashboards ({token.allowedDashboards.length})</p>
                  <div className="flex flex-wrap gap-1">
                    {token.allowedDashboards.map(uid => (
                      <span key={uid} className="text-[10px] bg-slate-700 text-slate-300 px-1.5 py-0.5 rounded">
                        {MOCK_DASHBOARDS.find(d => d.uid === uid)?.title || uid}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-slate-800 border border-slate-700 rounded-2xl w-full max-w-md overflow-hidden shadow-2xl">
            <div className="p-6 border-b border-slate-700 flex justify-between items-center">
              <h3 className="text-lg font-bold text-white">Create New Access Token</h3>
              <button onClick={() => setShowModal(false)} className="text-slate-400 hover:text-white">&times;</button>
            </div>
            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1">Display Name</label>
                <input 
                  type="text" required 
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white outline-none focus:border-blue-500"
                  placeholder="e.g. CI/CD Pipeline"
                  value={formData.name}
                  onChange={e => setFormData({...formData, name: e.target.value})}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1">Team</label>
                <input 
                  type="text" required
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white outline-none focus:border-blue-500"
                  placeholder="e.g. Infrastructure"
                  value={formData.team}
                  onChange={e => setFormData({...formData, team: e.target.value})}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1">Rate Limit (req/min)</label>
                <input 
                  type="number" required
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white outline-none focus:border-blue-500"
                  value={formData.rateLimit}
                  onChange={e => setFormData({...formData, rateLimit: parseInt(e.target.value)})}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Allowed Dashboards</label>
                <div className="space-y-2 max-h-40 overflow-y-auto pr-2 custom-scrollbar">
                  {MOCK_DASHBOARDS.map(d => (
                    <label key={d.uid} className="flex items-center gap-3 bg-slate-900/50 p-2 rounded border border-slate-700/50 cursor-pointer hover:bg-slate-700/50">
                      <input 
                        type="checkbox" 
                        checked={formData.allowedDashboards.includes(d.uid)}
                        onChange={() => toggleDashboard(d.uid)}
                        className="rounded border-slate-700 bg-slate-900 text-blue-600 focus:ring-blue-500"
                      />
                      <div className="text-sm">
                        <div className="text-slate-200 font-medium">{d.title}</div>
                        <div className="text-xs text-slate-500">{d.folderTitle}</div>
                      </div>
                    </label>
                  ))}
                </div>
              </div>
              <div className="pt-4 flex gap-3">
                <button type="button" onClick={() => setShowModal(false)} className="flex-1 px-4 py-2 border border-slate-700 text-slate-300 rounded-lg font-medium hover:bg-slate-700 transition-colors">Cancel</button>
                <button type="submit" className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-500 transition-colors shadow-lg shadow-blue-500/20">Create Token</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
