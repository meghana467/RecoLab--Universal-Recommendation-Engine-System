'use client';

import { useState } from 'react';
import { api, Recommendation } from '@/lib/api';
import RecommendationCard from '@/components/RecommendationCard';

const DEMO_USERS = [
  { id: 'u_ai_reader',    label: 'AI Reader',    emoji: '🤖', color: 'bg-purple-100 text-purple-800 border-purple-200' },
  { id: 'u_trade_reader', label: 'Trade Reader',  emoji: '📦', color: 'bg-teal-100 text-teal-800 border-teal-200' },
  { id: 'u_sports_reader',label: 'Sports Reader', emoji: '⚽', color: 'bg-yellow-100 text-yellow-800 border-yellow-200' },
];

function UserProductCard({ user }: { user: typeof DEMO_USERS[0] }) {
  const [recs, setRecs] = useState<Recommendation[]>([]);
  const [interests, setInterests] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const d = await api.getProductRecommendations(user.id);
      setRecs(d.recommendations || []);
      setInterests((d as { user_interests?: string[] }).user_interests || []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Error');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex-1 min-w-0 bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
      {/* Header */}
      <div className={`px-4 py-3 border-b ${user.color} border-opacity-50`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-xl">{user.emoji}</span>
            <div>
              <p className="text-sm font-semibold">{user.label}</p>
              <p className="text-xs opacity-70 font-mono">{user.id}</p>
            </div>
          </div>
          <button
            onClick={load}
            disabled={loading}
            className="px-3 py-1 bg-white bg-opacity-60 border border-current border-opacity-30 rounded-lg text-xs font-medium hover:bg-opacity-90 transition disabled:opacity-50"
          >
            {loading ? '...' : 'Load'}
          </button>
        </div>

        {interests.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {interests.map((l) => (
              <span key={l} className="text-xs bg-white bg-opacity-50 px-2 py-0.5 rounded-full font-medium">
                {l}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Body */}
      <div className="p-3 space-y-2 min-h-[200px]">
        {error && <p className="text-xs text-red-500">{error}</p>}

        {loading ? (
          [...Array(5)].map((_, i) => (
            <div key={i} className="bg-gray-100 rounded-lg h-14 animate-pulse" />
          ))
        ) : recs.length > 0 ? (
          recs.map((r, i) => (
            <RecommendationCard key={r.item_id} rec={r} rank={i + 1} showEvents userId={user.id} />
          ))
        ) : (
          <div className="text-center py-10 text-gray-400 text-sm">
            Click Load to get product recommendations
          </div>
        )}
      </div>
    </div>
  );
}

export default function ProductsPage() {
  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">E-commerce Product Recommendations</h1>
        <p className="text-sm text-gray-500 mt-1">
          Each user sees products matching their interest profile
        </p>
      </div>

      <div className="flex gap-4 overflow-x-auto pb-2">
        {DEMO_USERS.map((u) => (
          <UserProductCard key={u.id} user={u} />
        ))}
      </div>
    </div>
  );
}
