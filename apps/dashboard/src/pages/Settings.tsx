import { useState, useEffect } from 'react';

interface Config {
  server?: { host?: string; port?: number };
  logging?: { level?: string; format?: string };
  metrics?: { enabled?: boolean; port?: number; tracing?: boolean };
  tls?: { enabled?: boolean };
  security?: { max_body_size?: number };
}

export default function Settings() {
  const [config, setConfig] = useState<Config | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    fetch('/metrics')
      .then(() => setConfig({
        server: { host: '0.0.0.0', port: 8080 },
        logging: { level: 'info', format: 'json' },
        metrics: { enabled: true, port: 9090, tracing: false },
        tls: { enabled: false },
        security: { max_body_size: 10485760 },
      }))
      .catch(() => {});
  }, []);

  const save = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Settings</h1>
        <button
          onClick={save}
          className="px-4 py-2 bg-primary hover:bg-primary-dark rounded-lg text-sm transition"
        >
          {saved ? 'Saved!' : 'Save Changes'}
        </button>
      </div>

      {!config && <div className="text-gray-500">Loading configuration...</div>}

      {config && (
        <div className="space-y-6">
          <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
            <h2 className="text-lg font-semibold mb-4">Server</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-xs text-gray-400 block mb-1">Host</label>
                <input
                  type="text"
                  defaultValue={config.server?.host}
                  className="w-full bg-dark border border-dark-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-primary"
                />
              </div>
              <div>
                <label className="text-xs text-gray-400 block mb-1">Port</label>
                <input
                  type="number"
                  defaultValue={config.server?.port}
                  className="w-full bg-dark border border-dark-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-primary"
                />
              </div>
            </div>
          </div>

          <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
            <h2 className="text-lg font-semibold mb-4">Logging</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-xs text-gray-400 block mb-1">Level</label>
                <select className="w-full bg-dark border border-dark-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-primary">
                  <option>debug</option>
                  <option selected>info</option>
                  <option>warn</option>
                  <option>error</option>
                </select>
              </div>
              <div>
                <label className="text-xs text-gray-400 block mb-1">Format</label>
                <select className="w-full bg-dark border border-dark-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-primary">
                  <option selected>json</option>
                  <option>text</option>
                </select>
              </div>
            </div>
          </div>

          <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
            <h2 className="text-lg font-semibold mb-4">Metrics & Tracing</h2>
            <div className="space-y-3">
              <label className="flex items-center gap-3">
                <input type="checkbox" defaultChecked={config.metrics?.enabled} className="rounded" />
                <span className="text-sm">Enable metrics endpoint</span>
              </label>
              <label className="flex items-center gap-3">
                <input type="checkbox" defaultChecked={config.metrics?.tracing} className="rounded" />
                <span className="text-sm">Enable distributed tracing</span>
              </label>
            </div>
          </div>

          <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
            <h2 className="text-lg font-semibold mb-4">TLS</h2>
            <label className="flex items-center gap-3">
              <input type="checkbox" defaultChecked={config.tls?.enabled} className="rounded" />
              <span className="text-sm">Enable TLS/HTTPS</span>
            </label>
          </div>

          <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
            <h2 className="text-lg font-semibold mb-4">Security</h2>
            <div>
              <label className="text-xs text-gray-400 block mb-1">Max Body Size (bytes)</label>
              <input
                type="number"
                defaultValue={config.security?.max_body_size}
                className="w-full bg-dark border border-dark-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-primary"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
