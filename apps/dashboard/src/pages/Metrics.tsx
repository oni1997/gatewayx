import { useState, useEffect } from 'react';
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from 'recharts';

interface Snapshot {
  requests: number;
  active_connections: number;
  bytes_sent: number;
  bytes_received: number;
  routes: Array<{ name: string; count: number; avg_ms: number }>;
}

export default function Metrics() {
  const [history, setHistory] = useState<Snapshot[]>([]);
  const [current, setCurrent] = useState<Snapshot | null>(null);

  useEffect(() => {
    const fetchMetrics = () => {
      fetch('/metrics')
        .then((r) => r.json())
        .then((data) => {
          setCurrent(data);
          setHistory((prev) => [...prev.slice(-30), data]);
        })
        .catch(() => {});
    };

    fetchMetrics();
    const interval = setInterval(fetchMetrics, 5000);
    return () => clearInterval(interval);
  }, []);

  const chartData = history.map((s, i) => ({
    time: i,
    requests: s.requests,
    active: s.active_connections,
  }));

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Metrics</h1>

      <div className="bg-dark-card rounded-xl p-5 border border-dark-border mb-6">
        <h3 className="text-sm text-gray-400 mb-3">Requests Over Time</h3>
        <ResponsiveContainer width="100%" height={250}>
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
            <XAxis dataKey="time" stroke="#64748b" />
            <YAxis stroke="#64748b" />
            <Tooltip
              contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px' }}
              labelStyle={{ color: '#94a3b8' }}
            />
            <Line type="monotone" dataKey="requests" stroke="#2563eb" strokeWidth={2} dot={false} />
            <Line type="monotone" dataKey="active" stroke="#22c55e" strokeWidth={2} dot={false} />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {current && (
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
            <h3 className="text-xs text-gray-500 mb-2">Memory</h3>
            <div className="text-sm text-gray-400">
              Check the Prometheus endpoint at <code className="text-primary bg-dark rounded px-1">:9090/metrics</code>
            </div>
          </div>
          <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
            <h3 className="text-xs text-gray-500 mb-2">Goroutines</h3>
            <div className="text-sm text-gray-400">
              Check the Prometheus endpoint at <code className="text-primary bg-dark rounded px-1">:9090/metrics</code>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
