
import React from 'react';

interface StatCardProps {
  label: string;
  value: string | number;
  subValue?: string;
  icon: React.ReactNode;
  colorClass?: string;
}

export const StatCard: React.FC<StatCardProps> = ({ label, value, subValue, icon, colorClass = "text-blue-400" }) => {
  return (
    <div className="bg-slate-800 border border-slate-700 p-6 rounded-xl shadow-lg">
      <div className="flex items-center justify-between mb-4">
        <span className="text-slate-400 font-medium text-sm uppercase tracking-wider">{label}</span>
        <div className={`${colorClass} p-2 bg-slate-900/50 rounded-lg`}>
          {icon}
        </div>
      </div>
      <div className="flex items-baseline gap-2">
        <h3 className="text-3xl font-bold text-white">{value}</h3>
        {subValue && <span className="text-xs text-slate-500">{subValue}</span>}
      </div>
    </div>
  );
};
