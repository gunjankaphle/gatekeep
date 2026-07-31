import { NavLink } from 'react-router-dom';
import { LayoutDashboard, FileText, Network, GitCompare } from 'lucide-react';
import { cn } from '@/lib/utils';

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Logs', href: '/logs', icon: FileText },
  { name: 'Roles', href: '/roles', icon: Network },
  { name: 'Diff', href: '/diff', icon: GitCompare },
];

export function Sidebar() {
  return (
    <aside className="w-64 border-r border-slate-200 bg-slate-50">
      <nav className="space-y-1 p-4">
        {navigation.map((item) => (
          <NavLink
            key={item.name}
            to={item.href}
            end={item.href === '/'}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-slate-900 text-white'
                  : 'text-slate-700 hover:bg-slate-200'
              )
            }
          >
            <item.icon className="h-5 w-5" />
            {item.name}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
