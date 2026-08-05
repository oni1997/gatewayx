import { useState } from 'react';

interface APIKey {
  id: string;
  name: string;
  key: string;
  owner: string;
  created: string;
  lastUsed: string;
}

const mockKeys: APIKey[] = [
  { id: '1', name: 'Production API', key: 'sk-prod-' + 'x'.repeat(24), owner: 'admin', created: '2026-06-15', lastUsed: '2026-08-04' },
  { id: '2', name: 'Staging API', key: 'sk-stag-' + 'x'.repeat(24), owner: 'dev-team', created: '2026-07-01', lastUsed: '2026-08-03' },
  { id: '3', name: 'Mobile App', key: 'sk-mob-' + 'x'.repeat(24), owner: 'mobile', created: '2026-07-20', lastUsed: '2026-08-04' },
];

export default function APIKeys() {
  const [keys] = useState<APIKey[]>(mockKeys);
  const [showNew, setShowNew] = useState(false);
  const [newKey, setNewKey] = useState({ name: '', owner: '' });
  const [generated, setGenerated] = useState('');

  const generate = () => {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let key = 'sk-' + Array.from({ length: 32 }, () => chars[Math.floor(Math.random() * chars.length)]).join('');
    setGenerated(key);
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">API Keys</h1>
        <button
          onClick={() => setShowNew(!showNew)}
          className="px-4 py-2 bg-primary hover:bg-primary-dark rounded-lg text-sm transition"
        >
          + Generate Key
        </button>
      </div>

      {showNew && (
        <div className="bg-dark-card rounded-xl p-5 border border-dark-border mb-6">
          <h3 className="font-semibold mb-3">Generate New API Key</h3>
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className="text-xs text-gray-400 block mb-1">Name</label>
              <input
                type="text"
                value={newKey.name}
                onChange={(e) => setNewKey({ ...newKey, name: e.target.value })}
                placeholder="e.g. Production API"
                className="w-full bg-dark border border-dark-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-primary"
              />
            </div>
            <div>
              <label className="text-xs text-gray-400 block mb-1">Owner</label>
              <input
                type="text"
                value={newKey.owner}
                onChange={(e) => setNewKey({ ...newKey, owner: e.target.value })}
                placeholder="e.g. admin"
                className="w-full bg-dark border border-dark-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-primary"
              />
            </div>
          </div>
          {!generated ? (
            <button onClick={generate} className="px-4 py-2 bg-green-600 hover:bg-green-700 rounded-lg text-sm">
              Generate
            </button>
          ) : (
            <div>
              <div className="bg-dark rounded-lg p-3 font-mono text-xs text-green-400 mb-2 break-all">{generated}</div>
              <span className="text-xs text-yellow-400">Copy this key now. You won't be able to see it again.</span>
            </div>
          )}
        </div>
      )}

      <div className="bg-dark-card rounded-xl border border-dark-border overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-dark-border text-gray-400">
              <th className="text-left p-3">Name</th>
              <th className="text-left p-3">Owner</th>
              <th className="text-left p-3">Key</th>
              <th className="text-right p-3">Created</th>
              <th className="text-right p-3">Last Used</th>
              <th className="text-right p-3">Actions</th>
            </tr>
          </thead>
          <tbody>
            {keys.map((key) => (
              <tr key={key.id} className="border-b border-dark-border/50">
                <td className="p-3 font-medium">{key.name}</td>
                <td className="p-3">{key.owner}</td>
                <td className="p-3 font-mono text-xs text-gray-500">
                  {key.key.slice(0, 12)}...
                </td>
                <td className="p-3 text-right text-xs">{key.created}</td>
                <td className="p-3 text-right text-xs">{key.lastUsed}</td>
                <td className="p-3 text-right">
                  <button className="text-red-400 hover:text-red-300 text-xs">Revoke</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
