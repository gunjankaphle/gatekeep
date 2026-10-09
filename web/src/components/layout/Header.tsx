import { Shield } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';

export function Header() {
  const health = useQuery({ queryKey: ["health"], queryFn: () => api.getHealth(), refetchInterval: 30000 });
  return (
    <header className="sticky top-0 z-40 border-b border-slate-200 bg-white">
      <div className="flex h-16 items-center justify-between px-6">
        <div className="flex items-center gap-3">
          <Shield className="h-8 w-8 text-slate-900" />
          <div>
            <h1 className="text-xl font-bold text-slate-900">GateKeep</h1>
            <p className="text-xs text-slate-500">Snowflake Permission Manager</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <div className="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-1.5">
            <div className={`h-2 w-2 rounded-full ${health.data?.status === "healthy" ? "bg-green-500" : "bg-amber-500"}`} />
            <span className="text-sm text-slate-600">{health.isPending ? "Connecting…" : health.isError ? "API unavailable" : `API ${health.data?.status}`}</span>
          </div>
        </div>
      </div>
    </header>
  );
}
