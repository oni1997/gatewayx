import { useState } from 'react';

interface CertInfo {
  id: string;
  domain: string;
  issuer: string;
  expires: string;
  status: 'active' | 'expiring' | 'expired';
}

const mockCerts: CertInfo[] = [
  { id: '1', domain: 'api.example.com', issuer: "Let's Encrypt", expires: '2026-10-15', status: 'active' },
  { id: '2', domain: 'admin.example.com', issuer: "Let's Encrypt", expires: '2026-08-30', status: 'expiring' },
  { id: '3', domain: '*.internal.local', issuer: 'Self-Signed', expires: '2027-01-01', status: 'active' },
];

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  expiring: 'bg-yellow-500/20 text-yellow-400',
  expired: 'bg-red-500/20 text-red-400',
};

export default function Certificates() {
  const [certs] = useState<CertInfo[]>(mockCerts);
  const [renewing, setRenewing] = useState<string | null>(null);

  const renew = (domain: string) => {
    setRenewing(domain);
    setTimeout(() => setRenewing(null), 3000);
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Certificates</h1>
        <button className="px-4 py-2 bg-primary hover:bg-primary-dark rounded-lg text-sm transition">
          + Request Certificate
        </button>
      </div>

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
          <span className="text-sm text-gray-400">Issuer</span>
          <div className="text-lg font-semibold mt-1">Let's Encrypt</div>
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
              <th className="text-right p-3">Actions</th>
            </tr>
          </thead>
          <tbody>
            {certs.map((cert) => (
              <tr key={cert.id} className="border-b border-dark-border/50">
                <td className="p-3 font-mono text-xs">{cert.domain}</td>
                <td className="p-3">{cert.issuer}</td>
                <td className="p-3 text-right text-xs">{cert.expires}</td>
                <td className="p-3 text-right">
                  <span className={`px-2 py-0.5 rounded-full text-xs ${statusColors[cert.status]}`}>
                    {cert.status}
                  </span>
                </td>
                <td className="p-3 text-right">
                  <button
                    onClick={() => renew(cert.domain)}
                    disabled={renewing === cert.domain}
                    className="text-primary hover:text-primary-dark text-xs disabled:text-gray-500"
                  >
                    {renewing === cert.domain ? 'Renewing...' : 'Renew'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
