'use client';

interface FilterEffectivenessProps {
  data: any[];
}

export function FilterEffectiveness({ data }: FilterEffectivenessProps) {
  if (!data || data.length === 0) {
    return (
      <div className="text-slate-400 text-center py-8">
        No filter data available
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-slate-200">
            <th className="text-left py-3 px-4 font-semibold text-slate-900">Filter Name</th>
            <th className="text-right py-3 px-4 font-semibold text-slate-900">Usage Count</th>
            <th className="text-right py-3 px-4 font-semibold text-slate-900">Effectiveness</th>
            <th className="text-right py-3 px-4 font-semibold text-slate-900">Tokens Saved</th>
          </tr>
        </thead>
        <tbody>
          {data.map((filter) => (
            <tr key={filter.filterName} className="border-b border-slate-100 hover:bg-slate-50">
              <td className="py-3 px-4 text-slate-900 font-medium">{filter.filterName}</td>
              <td className="text-right py-3 px-4 text-slate-600">{filter.usageCount}</td>
              <td className="text-right py-3 px-4">
                <div className="flex items-center justify-end gap-2">
                  <div className="w-20 h-2 bg-slate-200 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-tokman-600"
                      style={{
                        width: `${Math.min(filter.avgEffectiveness * 100, 100)}%`,
                      }}
                    />
                  </div>
                  <span className="text-slate-900 font-semibold">
                    {(filter.avgEffectiveness * 100).toFixed(1)}%
                  </span>
                </div>
              </td>
              <td className="text-right py-3 px-4 text-green-600 font-semibold">
                {(filter.totalTokensSaved / 1000).toFixed(1)}K
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
