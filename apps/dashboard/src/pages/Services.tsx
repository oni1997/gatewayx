import { useState, useEffect } from 'react';
import StatCard from '../components/StatCard';

interface Route {
  name: string;
  count: number;
  avg_ms: number;
  min_ms: number;
  max_ms: number;
  last_ms: number;
}

export default function Services() {
  const [routes, setRoutes] = useState<Route[]>([]);

  useEffect(() => {
    fetch('/metrics')
      .then((r) => r.json())
      .then((data) => setRoutes(data.routes || []))
      .catch(() => {});
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Services</h1>

      <div className="grid grid-cols-3 gap-4 mb-6">
        <StatCard title="Active Routes" value={routes.length} />
        <StatCard title="Total Requests" value={routes.reduce((s, r) => s + r.count, 0).toLocaleString()} />
        <StatCard
          title="Avg Latency"
          value={
            routes.length > 0
              ? `${(routes.reduce((s, r) => s + r.avg_ms, 0) / routes.length).toFixed(1)} ms`
              : '-'
          }
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        {routes.map((route) => (
          <div key={route.name} className="bg-dark-card rounded-xl p-5 border border-dark-border">
            <h3 className="font-mono text-sm text-primary mb-3">{route.name}</h3>
            <div className="grid grid-cols-3 gap-3 text-xs">
              <div>
                <span className="text-gray-500">Requests</span>
                <p className="font-semibold mt-0.5">{route.count.toLocaleString()}</p>
              </div>
              <div>
                <span className="text-gray-500">Avg</span>
                <p className="font-semibold mt-0.5">{route.avg_ms.toFixed(1)} ms</p>
              </div>
              <div>
                <span className="text-gray-500">Max</span>
                <p className="font-semibold mt-0.5">{route.max_ms.toFixed(1)} ms</p>
              </div>
              <div>
                <span className="text-gray-500">Min</span>
                <p className="font-semibold mt-0.5">{route.min_ms.toFixed(1)} ms</p>
              </div>
              <div>
                <span className="text-gray-500">Last</span>
                <p className="font-semibold mt-0.5">{route.last_ms.toFixed(1)} ms</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {routes.length === 0 && (
        <div className="text-center text-gray-500 py-12">No services registered yet</div>
      )}
    </div>
  );
}
