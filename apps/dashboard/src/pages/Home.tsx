import { useState, useEffect } from 'react';
import StatCard from '../components/StatCard';
import { ChartIcon, ServerIcon, ClockIcon } from '../components/Icons';

interface Snapshot {
  requests: number;
  active_connections: number;
  bytes_sent: number;
  bytes_received: number;
  routes: Array<{ name: string; count: number; avg_ms: number }>;
}

export default function Home() {
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null);
  const [health, setHealth] = useState<{ status: string } | null>(null);

  useEffect(() => {
    fetch('/metrics')
      .then((r) => r.json())
      .then(setSnapshot)
      .catch(() => {});
    fetch('/health')
      .then((r) => r.json())
      .then(setHealth)
      .catch(() => {});
  }, []);

  const status = health?.status === 'healthy' ? 'healthy' : health?.status === 'degraded' ? 'degraded' : 'unhealthy';

  const topRoutes = (snapshot?.routes || [])
    .sort((a, b) => b.count - a.count)
    .slice(0, 5);

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-semibold text-white tracking-tight">Overview</h1>
        <p className="text-sm text-gray-500 mt-1">Gateway performance at a glance.</p>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard title="Total Requests" value={snapshot?.requests?.toLocaleString() ?? '-'} status={status} icon={<ChartIcon className="w-4 h-4" />} />
        <StatCard title="Active Connections" value={snapshot?.active_connections ?? '-'} icon={<ServerIcon className="w-4 h-4" />} />
        <StatCard title="Data Sent" value={snapshot?.bytes_sent ? `${(snapshot.bytes_sent / 1024 / 1024).toFixed(1)} MB` : '-'} />
        <StatCard title="Data Received" value={snapshot?.bytes_received ? `${(snapshot.bytes_received / 1024 / 1024).toFixed(1)} MB` : '-'} />
      </div>

      <div className="mb-3 flex items-center gap-2">
        <ClockIcon className="w-4 h-4 text-gray-500" />
        <h2 className="text-[15px] font-semibold text-white">Top Routes</h2>
      </div>
      <div className="bg-dark-card rounded-2xl border border-dark-border overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-dark-border text-gray-500 text-xs">
              <th className="text-left p-4 font-medium">Route</th>
              <th className="text-right p-4 font-medium">Requests</th>
              <th className="text-right p-4 font-medium">Avg Latency</th>
            </tr>
          </thead>
          <tbody>
            {topRoutes.map((route) => (
              <tr key={route.name} className="border-b border-dark-border/50 last:border-0 hover:bg-dark-hover/50 transition-colors">
                <td className="p-4 font-mono text-xs text-gray-300">{route.name}</td>
                <td className="p-4 text-right text-gray-200">{route.count.toLocaleString()}</td>
                <td className="p-4 text-right text-gray-400">{route.avg_ms.toFixed(1)} ms</td>
              </tr>
            ))}
            {topRoutes.length === 0 && (
              <tr>
                <td colSpan={3} className="p-4 text-center text-gray-500">No requests recorded yet</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
