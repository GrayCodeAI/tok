'use client';

import { TrendingUp, DollarSign, Users, Zap } from 'lucide-react';

interface MetricsCardsProps {
  data: any;
}

export function MetricsCards({ data }: MetricsCardsProps) {
  const cards = [
    {
      label: 'Total Commands',
      value: data?.teamStats?.totalCommands || 0,
      icon: Zap,
      color: 'bg-blue-50 text-blue-600',
    },
    {
      label: 'Tokens Saved',
      value: `${(data?.teamStats?.totalTokensSaved / 1000).toFixed(1)}K`,
      icon: TrendingUp,
      color: 'bg-green-50 text-green-600',
    },
    {
      label: 'Cost Savings',
      value: `$${data?.economics?.estimatedCostSaved?.toFixed(2) || '0.00'}`,
      icon: DollarSign,
      color: 'bg-emerald-50 text-emerald-600',
    },
    {
      label: 'Team Members',
      value: data?.teamStats?.memberCount || 0,
      icon: Users,
      color: 'bg-purple-50 text-purple-600',
    },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {cards.map((card) => {
        const Icon = card.icon;
        return (
          <div
            key={card.label}
            className="bg-white rounded-lg border border-slate-200 shadow-sm p-6"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">{card.label}</p>
                <p className="text-2xl font-bold text-slate-900 mt-2">{card.value}</p>
              </div>
              <div className={`p-3 rounded-lg ${card.color}`}>
                <Icon className="w-6 h-6" />
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
