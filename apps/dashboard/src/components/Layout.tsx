import { useState, useEffect } from 'react';
import { Outlet, NavLink } from 'react-router-dom';
import { HomeIcon, ServerIcon, ChartIcon, ClockIcon, KeyIcon, ShieldIcon, HeartIcon, GearIcon } from './Icons';

const links = [
  { to: '/', label: 'Overview', Icon: HomeIcon },
  { to: '/services', label: 'Services', Icon: ServerIcon },
  { to: '/metrics', label: 'Metrics', Icon: ChartIcon },
  { to: '/history', label: 'Request Log', Icon: ClockIcon },
  { to: '/api-keys', label: 'API Keys', Icon: KeyIcon },
  { to: '/certificates', label: 'Certificates', Icon: ShieldIcon },
  { to: '/health', label: 'Health', Icon: HeartIcon },
  { to: '/settings', label: 'Settings', Icon: GearIcon },
];

export default function Layout() {
  const [version, setVersion] = useState('');

  useEffect(() => {
    fetch('/version')
      .then((r) => r.json())
      .then((d) => setVersion(d.version))
      .catch(() => setVersion(''));
  }, []);

  return (
    <div className="flex h-screen bg-dark">
      <aside className="w-60 bg-dark-card border-r border-dark-border flex flex-col shrink-0">
        <div className="flex items-center gap-2.5 px-5 h-16 border-b border-dark-border">
          <div className="w-7 h-7 rounded-lg bg-primary flex items-center justify-center">
            <span className="text-white font-bold text-sm">G</span>
          </div>
          <span className="font-semibold text-[15px] tracking-tight text-white">GatewayX</span>
        </div>
        <nav className="flex flex-col gap-0.5 p-3 flex-1 overflow-y-auto">
          {links.map(({ to, label, Icon }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-lg text-[13px] font-medium transition-colors ${
                  isActive
                    ? 'bg-primary/10 text-primary-light'
                    : 'text-gray-400 hover:text-gray-200 hover:bg-dark-hover'
                }`
              }
            >
              <Icon className="w-[18px] h-[18px]" />
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="px-5 py-4 border-t border-dark-border text-[11px] text-gray-500">
          {version || 'dev'}
        </div>
      </aside>
      <main className="flex-1 overflow-auto">
        <div className="max-w-6xl mx-auto px-8 py-8">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
