import { useState, useEffect } from 'react';

interface RequestEntry {
  timestamp: string;
  trace_id: string;
  span_id: string;
  method: string;
  path: string;
  host: string;
  remote_addr: string;
  status: number;
  duration_ms: number;
  bytes_sent: number;
}

const methodColors: Record<string, string> = {
  GET: 'text-green-400',
  POST: 'text-blue-400',
  PUT: 'text-yellow-400',
  DELETE: 'text-red-400',
  PATCH: 'text-purple-400',
};

export default function RequestHistory() {
  const [entries, setEntries] = useState<RequestEntry[]>([]);

  useEffect(() => {
    fetch('/history')
      .then((r) => r.json())
      .then((data) => setEntries(data.reverse()))
      .catch(() => {});
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Request History</h1>

      {entries.length === 0 && (
        <div className="text-center text-gray-500 py-12">No requests recorded yet</div>
      )}

      <div className="space-y-2">
        {entries.map((entry, i) => (
          <div key={`${entry.trace_id}-${i}`} className="bg-dark-card rounded-lg p-4 border border-dark-border hover:border-dark-border/80 transition">
            <div className="flex items-center gap-3 text-sm">
              <span className={`font-mono font-bold ${methodColors[entry.method] || 'text-gray-400'}`}>
                {entry.method}
              </span>
              <span className="font-mono text-xs text-gray-300 truncate flex-1">{entry.path}</span>
              <span className={`text-xs ${entry.status < 400 ? 'text-green-400' : 'text-red-400'}`}>
                {entry.status}
              </span>
              <span className="text-xs text-gray-500">{entry.duration_ms.toFixed(1)} ms</span>
            </div>
            <div className="flex gap-4 mt-2 text-xs text-gray-600">
              <span>Host: {entry.host}</span>
              <span>IP: {entry.remote_addr}</span>
              <span className="font-mono">Trace: {entry.trace_id?.slice(0, 12)}</span>
              {entry.bytes_sent > 0 && <span>Sent: {(entry.bytes_sent / 1024).toFixed(1)} KB</span>}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
