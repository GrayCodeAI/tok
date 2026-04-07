'use client';

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';

interface TrendChartProps {
  data: any[];
}

export function TrendChart({ data }: TrendChartProps) {
  if (!data || data.length === 0) {
    return (
      <div className="h-64 flex items-center justify-center text-slate-400">
        No data available
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="timestamp" />
        <YAxis />
        <Tooltip
          formatter={(value: any) => {
            if (typeof value === 'number') {
              return value.toLocaleString();
            }
            return value;
          }}
        />
        <Legend />
        <Line
          type="monotone"
          dataKey="totalSavedTokens"
          stroke="#0284c7"
          name="Tokens Saved"
        />
        <Line
          type="monotone"
          dataKey="estimatedSavingsUsd"
          stroke="#10b981"
          name="Cost Savings ($)"
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
