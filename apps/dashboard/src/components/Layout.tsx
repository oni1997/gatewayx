import { Outlet, NavLink } from 'react-router-dom';

const links = [
  { to: '/', label: 'Home', icon: '●' },
  { to: '/services', label: 'Services', icon: '◇' },
  { to: '/metrics', label: 'Metrics', icon: '◆' },
  { to: '/history', label: 'History', icon: '◈' },
  { to: '/api-keys', label: 'API Keys', icon: '▣' },
  { to: '/certificates', label: 'Certs', icon: '◉' },
  { to: '/health', label: 'Health', icon: '○' },
  { to: '/settings', label: 'Settings', icon: '◎' },
];

export default function Layout() {
  return (
    <div className="flex h-screen">
      <aside className="w-60 bg-dark-card border-r border-dark-border p-4 flex flex-col">
        <div className="flex items-center gap-2 mb-8 px-2">
          <span className="text-xl font-bold text-primary">GatewayX</span>
        </div>
        <nav className="flex flex-col gap-1">
          {links.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              end={link.to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition ${
                  isActive
                    ? 'bg-primary/20 text-primary'
                    : 'text-gray-400 hover:text-gray-200 hover:bg-dark-border/50'
                }`
              }
            >
              <span>{link.icon}</span>
              {link.label}
            </NavLink>
          ))}
        </nav>
        <div className="mt-auto pt-4 border-t border-dark-border text-xs text-gray-500">
          v0.1.0
        </div>
      </aside>
      <main className="flex-1 overflow-auto p-6">
        <Outlet />
      </main>
    </div>
  );
}
