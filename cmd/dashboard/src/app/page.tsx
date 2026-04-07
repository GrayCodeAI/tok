'use client';

import { useState, useEffect } from 'react';
import { Dashboard } from '@/components/dashboard';
import { Header } from '@/components/header';
import { Sidebar } from '@/components/sidebar';

export default function Home() {
  const [teamId, setTeamId] = useState<string>('');

  useEffect(() => {
    // Get team ID from localStorage or URL params
    const storedTeamId = localStorage.getItem('tokman_team_id');
    if (storedTeamId) {
      setTeamId(storedTeamId);
    }
  }, []);

  if (!teamId) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-tokman-50">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-tokman-900 mb-4">TokMan</h1>
          <p className="text-tokman-600 mb-8">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-slate-50">
      <Sidebar />
      <div className="flex-1 flex flex-col overflow-hidden">
        <Header teamId={teamId} />
        <main className="flex-1 overflow-auto">
          <Dashboard teamId={teamId} />
        </main>
      </div>
    </div>
  );
}
