'use client';

import { BarChart3, Settings, Bell, User } from 'lucide-react';

export function Header({ teamId }: { teamId: string }) {
  return (
    <header className="border-b border-slate-200 bg-white px-6 py-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <BarChart3 className="w-8 h-8 text-tokman-600" />
          <h1 className="text-2xl font-bold text-slate-900">Analytics Dashboard</h1>
        </div>

        <div className="flex items-center gap-4">
          <button className="p-2 hover:bg-slate-100 rounded-lg transition-colors">
            <Bell className="w-5 h-5 text-slate-600" />
          </button>

          <button className="p-2 hover:bg-slate-100 rounded-lg transition-colors">
            <Settings className="w-5 h-5 text-slate-600" />
          </button>

          <button className="p-2 hover:bg-slate-100 rounded-lg transition-colors">
            <User className="w-5 h-5 text-slate-600" />
          </button>
        </div>
      </div>

      <div className="mt-4 text-sm text-slate-600">
        Team ID: <span className="font-mono text-tokman-600">{teamId}</span>
      </div>
    </header>
  );
}
