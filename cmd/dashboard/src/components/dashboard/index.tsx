'use client';

import { useDashboardData } from '@/hooks/use-dashboard';
import { MetricsCards } from './metrics-cards';
import { TrendChart } from './trend-chart';
import { FilterEffectiveness } from './filter-effectiveness';
import { CostProjection } from './cost-projection';
import { LoadingSpinner } from '@/components/loading-spinner';

interface DashboardProps {
  teamId: string;
}

export function Dashboard({ teamId }: DashboardProps) {
  const { data, isLoading, error } = useDashboardData(teamId);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h3 className="text-red-900 font-semibold">Error loading dashboard</h3>
          <p className="text-red-700 text-sm">{error}</p>
        </div>
      </div>
    );
  }

  if (!data) {
    return null;
  }

  return (
    <div className="p-6 space-y-6">
      {/* Metrics Cards */}
      <MetricsCards data={data} />

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Trend Chart */}
        <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-6">
          <h2 className="text-lg font-semibold text-slate-900 mb-4">Token Savings Trend</h2>
          <TrendChart data={data.trends} />
        </div>

        {/* Cost Projection */}
        <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-6">
          <h2 className="text-lg font-semibold text-slate-900 mb-4">Cost Projection</h2>
          <CostProjection data={data.projection} />
        </div>
      </div>

      {/* Filter Effectiveness */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-6">
        <h2 className="text-lg font-semibold text-slate-900 mb-4">Top Performing Filters</h2>
        <FilterEffectiveness data={data.topFilters} />
      </div>
    </div>
  );
}
