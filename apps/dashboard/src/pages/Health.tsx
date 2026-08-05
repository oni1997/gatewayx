import { useState, useEffect } from 'react';

interface HealthReport {
  status: string;
  uptime: string;
  checks: Record<string, string>;
  timestamp: string;
}

export default function Health() {
  const [health, setHealth] = useState<HealthReport | null>(null);

  useEffect(() => {
    fetch('/health')
      .then((r) => r.json())
      .then(setHealth)
      .catch(() => {});
  }, []);

  const statusColors: Record<string, string> = {
    healthy: 'text-green-400',
    degraded: 'text-yellow-400',
    unhealthy: 'text-red-400',
  };

  const statusDots: Record<string, string> = {
    healthy: 'bg-green-500',
    degraded: 'bg-yellow-500',
    unhealthy: 'bg-red-500',
  };

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Health</h1>

      {health && (
        <div className="space-y-4">
          <div className="grid grid-cols-3 gap-4">
            <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
              <span className="text-sm text-gray-400">Status</span>
              <div className={`text-xl font-semibold capitalize mt-1 ${statusColors[health.status] || ''}`}>
                {health.status}
              </div>
            </div>
            <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
              <span className="text-sm text-gray-400">Uptime</span>
              <div className="text-xl font-semibold mt-1">{health.uptime}</div>
            </div>
            <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
              <span className="text-sm text-gray-400">Last Check</span>
              <div className="text-xl font-semibold mt-1">
                {new Date(health.timestamp).toLocaleTimeString()}
              </div>
            </div>
          </div>

          <h2 className="text-lg font-semibold mt-6 mb-3">Service Checks</h2>
          <div className="bg-dark-card rounded-xl border border-dark-border overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-dark-border text-gray-400">
                  <th className="text-left p-3">Service</th>
                  <th className="text-right p-3">Status</th>
                </tr>
              </thead>
              <tbody>
                {Object.entries(health.checks).map(([name, status]) => (
                  <tr key={name} className="border-b border-dark-border/50">
                    <td className="p-3 font-mono text-xs">{name}</td>
                    <td className="p-3 text-right">
                      <span className="flex items-center justify-end gap-2">
                        <span className={`w-2 h-2 rounded-full ${statusDots[status === 'healthy' ? 'healthy' : 'unhealthy']}`} />
                        {status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
