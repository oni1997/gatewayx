interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  status?: 'healthy' | 'degraded' | 'unhealthy';
}

export default function StatCard({ title, value, subtitle, status }: StatCardProps) {
  const statusColors: Record<string, string> = {
    healthy: 'bg-green-500',
    degraded: 'bg-yellow-500',
    unhealthy: 'bg-red-500',
  };

  return (
    <div className="bg-dark-card rounded-xl p-5 border border-dark-border">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm text-gray-400">{title}</span>
        {status && (
          <span className={`w-2 h-2 rounded-full ${statusColors[status] || 'bg-gray-500'}`} />
        )}
      </div>
      <div className="text-2xl font-semibold">{value}</div>
      {subtitle && <div className="text-xs text-gray-500 mt-1">{subtitle}</div>}
    </div>
  );
}
