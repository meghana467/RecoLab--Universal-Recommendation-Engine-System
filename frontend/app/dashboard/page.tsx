'use client';

import { useEffect, useState } from 'react';
import { api, DashboardStats } from '@/lib/api';

function StatCard({ label, value, icon, color }: { label: string; value: number | string; icon: string; color: string }) {
  return (
    <div className={`bg-white rounded-xl border border-gray-200 p-5 shadow-sm`}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-500">{label}</p>
          <p className={`text-3xl font-bold mt-1 ${color}`}>{typeof value === 'number' ? value.toLocaleString() : value}</p>
        </div>
        <span className="text-4xl">{icon}</span>
      </div>
    </div>
  );
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [seeding, setSeeding] = useState(false);
  const [seedResult, setSeedResult] = useState<string | null>(null);

  async function loadStats() {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getDashboard();
      setStats(data);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to load stats');
    } finally {
      setLoading(false);
    }
  }

  async function handleSeed() {
    setSeeding(true);
    setSeedResult(null);
    try {
      const result = await api.loadData();
      setSeedResult(`Data seeded successfully. Items: ${result['total_items_loaded']}, Users: ${result['users_loaded']}, Feedback: ${result['feedback_loaded']}`);
      await loadStats();
    } catch (e: unknown) {
      setSeedResult(`Seed error: ${e instanceof Error ? e.message : 'unknown'}`);
    } finally {
      setSeeding(false);
    }
  }

  useEffect(() => { loadStats(); }, []);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-sm text-gray-500 mt-1">RecoLab recommendation engine overview</p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={handleSeed}
            disabled={seeding}
            className="px-4 py-2 bg-green-600 text-white rounded-lg text-sm font-medium hover:bg-green-700 disabled:opacity-50 transition-colors"
          >
            {seeding ? 'Seeding...' : 'Seed Demo Data'}
          </button>
          <button
            onClick={loadStats}
            className="px-4 py-2 bg-brand-600 text-white rounded-lg text-sm font-medium hover:bg-brand-700 transition-colors"
          >
            Refresh
          </button>
        </div>
      </div>

      {seedResult && (
        <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-800">
          {seedResult}
        </div>
      )}

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-800">
          {error} — Is the backend running on port 3001?
        </div>
      )}

      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          {[1,2,3].map(i => (
            <div key={i} className="bg-white rounded-xl border border-gray-200 p-5 h-24 animate-pulse" />
          ))}
        </div>
      ) : stats ? (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
            <StatCard label="Total Users" value={stats.totalUsers} icon="👤" color="text-blue-600" />
            <StatCard label="Total Items" value={stats.totalItems} icon="📦" color="text-purple-600" />
            <StatCard label="Feedback Events" value={stats.totalFeedback} icon="📈" color="text-green-600" />
          </div>

          <h2 className="text-lg font-semibold mb-3">Category Breakdown</h2>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
            {stats.topCategories.map((cat) => (
              <div key={cat.category} className="bg-white rounded-xl border border-gray-200 p-4 shadow-sm">
                <p className="text-sm text-gray-500">{cat.name}</p>
                <p className="text-2xl font-bold text-gray-800 mt-1">{cat.count}</p>
                <span className="text-xs text-gray-400 font-mono">{cat.category}</span>
              </div>
            ))}
          </div>
        </>
      ) : null}

    </div>
  );
}
