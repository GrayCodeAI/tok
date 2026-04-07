'use client';

import { BarChart3, TrendingUp, Users, Settings, LogOut } from 'lucide-react';
import Link from 'next/link';
import { useState } from 'react';

const menuItems = [
  { label: 'Overview', href: '/', icon: BarChart3 },
  { label: 'Trends', href: '/trends', icon: TrendingUp },
  { label: 'Team', href: '/team', icon: Users },
  { label: 'Settings', href: '/settings', icon: Settings },
];

export function Sidebar() {
  const [isOpen, setIsOpen] = useState(true);

  return (
    <aside className={`${isOpen ? 'w-64' : 'w-20'} bg-tokman-900 text-white transition-all duration-300 flex flex-col`}>
      {/* Logo */}
      <div className="p-6 border-b border-tokman-800">
        <div className="flex items-center gap-3">
          <BarChart3 className="w-8 h-8" />
          {isOpen && <span className="font-bold text-lg">TokMan</span>}
        </div>
      </div>

      {/* Menu */}
      <nav className="flex-1 p-4 space-y-2">
        {menuItems.map((item) => {
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              className="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-tokman-800 transition-colors"
            >
              <Icon className="w-5 h-5" />
              {isOpen && <span>{item.label}</span>}
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="border-t border-tokman-800 p-4">
        <button className="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-tokman-800 transition-colors w-full text-left">
          <LogOut className="w-5 h-5" />
          {isOpen && <span>Logout</span>}
        </button>
      </div>
    </aside>
  );
}
