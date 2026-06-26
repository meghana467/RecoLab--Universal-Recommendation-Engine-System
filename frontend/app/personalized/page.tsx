'use client';

import { useState } from 'react';
import { api, RecommendationResponse } from '@/lib/api';
import RecommendationCard from '@/components/RecommendationCard';

const DEMO_USERS = [
  { id: 'u_ai_reader',    label: 'AI Reader',    emoji: '🤖', color: 'bg-purple-100 text-purple-800 border-purple-200' },
  { id: 'u_trade_reader', label: 'Trade Reader',  emoji: '📦', color: 'bg-teal-100 text-teal-800 border-teal-200' },
  { id: 'u_sports_reader',label: 'Sports Reader', emoji: '⚽', color: 'bg-yellow-100 text-yellow-800 border-yellow-200' },
];

function UserFeedCard({ user }: { user: typeof DEMO_USERS[0] }) {
  const [result, setResult] = useState<RecommendationResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getPersonalizedNews(user.id);
      setResult(data);
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

        {/* Interest tags */}
        {result?.user_interests && result.user_interests.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {result.user_interests.map((l) => (
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
        ) : result ? (
          result.recommendations.length > 0 ? (
            result.recommendations.map((rec, idx) => (
              <RecommendationCard key={rec.item_id} rec={rec} rank={idx + 1} showEvents userId={user.id} />
            ))
          ) : (
            <p className="text-sm text-gray-400 text-center py-8">No recommendations yet. Seed data first.</p>
          )
        ) : (
          <div className="text-center py-10 text-gray-400 text-sm">Click Load to fetch personalized feed</div>
        )}
      </div>
    </div>
  );
}

export default function PersonalizedPage() {
  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Personalized News Feed</h1>
        <p className="text-sm text-gray-500 mt-1">
          Each user gets different recommendations based on their interest labels
        </p>
      </div>

      <div className="flex gap-4 overflow-x-auto pb-2">
        {DEMO_USERS.map((u) => (
          <UserFeedCard key={u.id} user={u} />
        ))}
      </div>
    </div>
  );
}
