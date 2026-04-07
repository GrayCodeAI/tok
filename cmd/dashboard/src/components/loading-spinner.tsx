'use client';

export function LoadingSpinner() {
  return (
    <div className="flex flex-col items-center justify-center gap-3">
      <div className="w-8 h-8 border-4 border-tokman-200 border-t-tokman-600 rounded-full animate-spin" />
      <p className="text-slate-600 text-sm">Loading analytics data...</p>
    </div>
  );
}
