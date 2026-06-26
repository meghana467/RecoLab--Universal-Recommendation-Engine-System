import type { Metadata } from 'next';
import Link from 'next/link';
import './globals.css';

export const metadata: Metadata = {
  title: 'RecoLab — Universal Recommendation Engine Demo',
  description: 'Powered by Gorse recommendation engine',
};

const navItems = [
  { href: '/dashboard',    label: 'Dashboard',       icon: '📊' },
  { href: '/trending',     label: 'Trending News',   icon: '🔥' },
  { href: '/personalized', label: 'Personalized Feed', icon: '🎯' },
  { href: '/products',     label: 'E-commerce',      icon: '🛍️' },
  { href: '/expo',         label: 'Expo & Trade',    icon: '🏭' },
  { href: '/simulator',   label: 'Event Simulator',  icon: '⚡' },
];

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="flex h-screen overflow-hidden">
        {/* Sidebar */}
        <aside className="w-60 flex-shrink-0 bg-gray-900 text-white flex flex-col">
          <div className="px-5 py-5 border-b border-gray-700">
            <div className="text-lg font-bold tracking-tight text-white">RecoLab</div>
            <div className="text-xs text-gray-400 mt-0.5">Recommendation Engine Demo</div>
          </div>
          <nav className="flex-1 py-4 space-y-0.5 px-2">
            {navItems.map(({ href, label, icon }) => (
              <Link
                key={href}
                href={href}
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-gray-300 hover:bg-gray-800 hover:text-white transition-colors"
              >
                <span>{icon}</span>
                <span>{label}</span>
              </Link>
            ))}
          </nav>
          <div className="px-4 py-3 border-t border-gray-700 text-xs text-gray-500">
            Powered by Gorse + Go + Next.js
          </div>
        </aside>

        {/* Main */}
        <main className="flex-1 overflow-y-auto bg-gray-50">
          <div className="max-w-6xl mx-auto px-6 py-6">
            {children}
          </div>
        </main>
      </body>
    </html>
  );
}
