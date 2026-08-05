import { useState, useEffect } from 'react';
import StatCard from '../components/StatCard';

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
      <h1 className="text-2xl font-bold mb-6">Overview</h1>

      <div className="grid grid-cols-4 gap-4 mb-8">
        <StatCard title="Total Requests" value={snapshot?.requests?.toLocaleString() ?? '-'} status={status} />
        <StatCard title="Active Connections" value={snapshot?.active_connections ?? '-'} />
        <StatCard title="Data Sent" value={snapshot?.bytes_sent ? `${(snapshot.bytes_sent / 1024 / 1024).toFixed(1)} MB` : '-'} />
        <StatCard title="Data Received" value={snapshot?.bytes_received ? `${(snapshot.bytes_received / 1024 / 1024).toFixed(1)} MB` : '-'} />
      </div>

      <h2 className="text-lg font-semibold mb-3">Top Routes</h2>
      <div className="bg-dark-card rounded-xl border border-dark-border overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-dark-border text-gray-400">
              <th className="text-left p-3">Route</th>
              <th className="text-right p-3">Requests</th>
              <th className="text-right p-3">Avg Latency</th>
            </tr>
          </thead>
          <tbody>
            {topRoutes.map((route) => (
              <tr key={route.name} className="border-b border-dark-border/50">
                <td className="p-3 font-mono text-xs">{route.name}</td>
                <td className="p-3 text-right">{route.count.toLocaleString()}</td>
                <td className="p-3 text-right">{route.avg_ms.toFixed(1)} ms</td>
              </tr>
            ))}
            {topRoutes.length === 0 && (
              <tr>
                <td colSpan={3} className="p-3 text-center text-gray-500">No requests yet</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
