import { useState, useEffect } from 'react';

interface CertInfo {
  id: string;
  domain: string;
  issuer: string;
  not_after: string;
  status: string;
  created_at: string;
}

export default function Certificates() {
  const [certs, setCerts] = useState<CertInfo[]>([]);
  const [showNew, setShowNew] = useState(false);
  const [newCert, setNewCert] = useState({ domain: '', issuer: "Let's Encrypt" });

  const fetchCerts = () => {
    fetch('/api/certs')
      .then((r) => r.json())
      .then(setCerts)
      .catch(() => {});
  };

  useEffect(() => { fetchCerts(); }, []);

  const request = () => {
    fetch('/api/certs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newCert),
    })
      .then((r) => r.json())
      .then(() => {
        setShowNew(false);
        setNewCert({ domain: '', issuer: "Let's Encrypt" });
        fetchCerts();
      })
      .catch(() => {});
  };

  const statusColors: Record<string, string> = {
    active: 'bg-green-500/20 text-green-400',
    expiring: 'bg-yellow-500/20 text-yellow-400',
    expired: 'bg-red-500/20 text-red-400',
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Certificates</h1>
        <button
          onClick={() => setShowNew(!showNew)}
          className="px-4 py-2 bg-primary hover:bg-primary-dark rounded-lg text-sm transition"
        >
          + Add Certificate
        </button>
      </div>

      {showNew && (
        <div className="bg-dark-card rounded-xl p-5 border border-dark-border mb-6">
          <h3 className="font-semibold mb-3">Add Certificate</h3>
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className="text-xs text-gray-400 block mb-1">Domain</label>
              <input
                type="text"
                value={newCert.domain}
                onChange={(e) => setNewCert({ ...newCert, domain: e.target.value })}
                placeholder="api.example.com"
                className="w-full bg-dark border border-dark-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-primary"
              />
            </div>
            <div>
              <label className="text-xs text-gray-400 block mb-1">Issuer</label>
              <input
                type="text"
                value={newCert.issuer}
                onChange={(e) => setNewCert({ ...newCert, issuer: e.target.value })}
                className="w-full bg-dark border border-dark-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-primary"
              />
            </div>
          </div>
          <button onClick={request} className="px-4 py-2 bg-green-600 hover:bg-green-700 rounded-lg text-sm" disabled={!newCert.domain}>
            Add
          </button>
        </div>
      )}

      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
          <span className="text-sm text-gray-400">Active</span>
          <div className="text-2xl font-semibold mt-1 text-green-400">
            {certs.filter((c) => c.status === 'active').length}
          </div>
        </div>
        <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
          <span className="text-sm text-gray-400">Expiring Soon</span>
          <div className="text-2xl font-semibold mt-1 text-yellow-400">
            {certs.filter((c) => c.status === 'expiring').length}
          </div>
        </div>
        <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
          <span className="text-sm text-gray-400">Total</span>
          <div className="text-2xl font-semibold mt-1">{certs.length}</div>
        </div>
      </div>

      <div className="bg-dark-card rounded-xl border border-dark-border overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-dark-border text-gray-400">
              <th className="text-left p-3">Domain</th>
              <th className="text-left p-3">Issuer</th>
              <th className="text-right p-3">Expires</th>
              <th className="text-right p-3">Status</th>
            </tr>
          </thead>
          <tbody>
            {certs.map((cert) => (
              <tr key={cert.id} className="border-b border-dark-border/50">
                <td className="p-3 font-mono text-xs">{cert.domain}</td>
                <td className="p-3">{cert.issuer}</td>
                <td className="p-3 text-right text-xs">{new Date(cert.not_after).toLocaleDateString()}</td>
                <td className="p-3 text-right">
                  <span className={`px-2 py-0.5 rounded-full text-xs ${statusColors[cert.status] || 'bg-gray-500/20 text-gray-400'}`}>
                    {cert.status}
                  </span>
                </td>
              </tr>
            ))}
            {certs.length === 0 && (
              <tr><td colSpan={4} className="p-3 text-center text-gray-500">No certificates yet</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
