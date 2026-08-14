interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  status?: 'healthy' | 'degraded' | 'unhealthy';
  icon?: React.ReactNode;
}

export default function StatCard({ title, value, subtitle, status, icon }: StatCardProps) {
  const statusColors: Record<string, string> = {
    healthy: 'bg-emerald-500',
    degraded: 'bg-amber-500',
    unhealthy: 'bg-red-500',
  };

  return (
    <div className="bg-dark-card rounded-2xl p-5 border border-dark-border hover:border-dark-hover transition-colors">
      <div className="flex items-center justify-between mb-3">
        <span className="text-[13px] font-medium text-gray-400">{title}</span>
        {icon && <span className="text-gray-500">{icon}</span>}
        {status && (
          <span className={`w-2 h-2 rounded-full ${statusColors[status] || 'bg-gray-500'}`} />
        )}
      </div>
      <div className="text-2xl font-semibold text-white tracking-tight">{value}</div>
      {subtitle && <div className="text-xs text-gray-500 mt-1.5">{subtitle}</div>}
    </div>
  );
}
