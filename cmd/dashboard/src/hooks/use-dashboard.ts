import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

interface DashboardResponse {
  teamStats: {
    totalCommands: number;
    totalTokensSaved: number;
    avgReductionPercent: number;
    memberCount: number;
    estimatedMonthlyCost: number;
  };
  economics: {
    tokensSaved: number;
    estimatedCostSaved: number;
    roiPercent: number;
    monthlyCommands: number;
  };
  trends: Array<{
    timestamp: string;
    totalSavedTokens: number;
    estimatedCostUsd: number;
    estimatedSavingsUsd: number;
  }>;
  topFilters: Array<{
    filterName: string;
    usageCount: number;
    avgEffectiveness: number;
    totalTokensSaved: number;
  }>;
  projection: {
    projectedMonthlyCost: number;
    projectedMonthlySavings: number;
    actualMonthlyCost: number;
    actualMonthlySavings: number;
  };
}

export function useDashboardData(teamId: string) {
  return useQuery<DashboardResponse, Error>({
    queryKey: ['dashboard', teamId],
    queryFn: async () => {
      const response = await axios.get(
        `${process.env.NEXT_PUBLIC_API_URL}/dashboard`,
        {
          params: { team_id: teamId },
        }
      );
      return response.data;
    },
    enabled: !!teamId,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}
