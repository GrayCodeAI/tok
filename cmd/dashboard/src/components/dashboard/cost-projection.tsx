'use client';

interface CostProjectionProps {
  data: {
    projectedMonthlyCost: number;
    projectedMonthlySavings: number;
    actualMonthlyCost: number;
    actualMonthlySavings: number;
  };
}

export function CostProjection({ data }: CostProjectionProps) {
  if (!data) {
    return (
      <div className="text-slate-400 text-center py-8">
        No projection data available
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="bg-tokman-50 rounded-lg p-4">
          <p className="text-sm text-slate-600 mb-1">Actual Monthly Cost</p>
          <p className="text-2xl font-bold text-tokman-900">
            ${data.actualMonthlyCost?.toFixed(2) || '0.00'}
          </p>
        </div>
        <div className="bg-green-50 rounded-lg p-4">
          <p className="text-sm text-slate-600 mb-1">Actual Monthly Savings</p>
          <p className="text-2xl font-bold text-green-600">
            ${data.actualMonthlySavings?.toFixed(2) || '0.00'}
          </p>
        </div>
      </div>

      <div className="border-t border-slate-200 pt-4">
        <p className="text-xs text-slate-500 uppercase font-semibold tracking-wide mb-3">
          Projection (30 days)
        </p>
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-slate-50 rounded-lg p-4">
            <p className="text-sm text-slate-600 mb-1">Projected Monthly Cost</p>
            <p className="text-2xl font-bold text-slate-900">
              ${data.projectedMonthlyCost?.toFixed(2) || '0.00'}
            </p>
          </div>
          <div className="bg-emerald-50 rounded-lg p-4">
            <p className="text-sm text-slate-600 mb-1">Projected Monthly Savings</p>
            <p className="text-2xl font-bold text-emerald-600">
              ${data.projectedMonthlySavings?.toFixed(2) || '0.00'}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
