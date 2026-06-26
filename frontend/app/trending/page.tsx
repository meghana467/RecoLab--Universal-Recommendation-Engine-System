'use client';

import { useEffect, useState, useCallback } from 'react';
import { api, Recommendation } from '@/lib/api';
import RecommendationCard from '@/components/RecommendationCard';

const CATEGORIES = ['All', 'Technology', 'AI', 'Business', 'Sports', 'Trade', 'India', 'World'];

export default function TrendingPage() {
  const [items, setItems] = useState<Recommendation[]>([]);
  const [filtered, setFiltered] = useState<Recommendation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [category, setCategory] = useState('All');
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getTrendingNews(20);
      setItems(data.recommendations || []);
      setLastRefresh(new Date());
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to fetch');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  // Auto-refresh every 30 seconds
  useEffect(() => {
    const timer = setInterval(load, 30_000);
    return () => clearInterval(timer);
  }, [load]);

  // Filter by category
  useEffect(() => {
    if (category === 'All') {
      setFiltered(items);
    } else {
      const cat = category.toLowerCase();
      setFiltered(items.filter((r) => r.labels?.some((l) => l.toLowerCase() === cat)));
    }
  }, [category, items]);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Trending News</h1>
          <p className="text-sm text-gray-500 mt-1">
            Popular articles · auto-refreshes every 30s · last: {lastRefresh.toLocaleTimeString()}
          </p>
        </div>
        <button
          onClick={load}
          className="px-4 py-2 bg-brand-600 text-white rounded-lg text-sm font-medium hover:bg-brand-700 transition-colors"
        >
          Refresh Now
        </button>
      </div>

      {/* Category filters */}
      <div className="flex flex-wrap gap-2 mb-5">
        {CATEGORIES.map((cat) => (
          <button
            key={cat}
            onClick={() => setCategory(cat)}
            className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
              category === cat
                ? 'bg-brand-600 text-white'
                : 'bg-white border border-gray-300 text-gray-700 hover:bg-gray-50'
            }`}
          >
            {cat}
          </button>
        ))}
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-800">
          {error}
        </div>
      )}

      {loading ? (
        <div className="space-y-3">
          {[...Array(8)].map((_, i) => (
            <div key={i} className="bg-white rounded-xl border border-gray-200 p-4 h-20 animate-pulse" />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          No trending articles found. Seed demo data from the Dashboard.
        </div>
      ) : (
        <div className="space-y-3">
          {filtered.map((rec, idx) => (
            <RecommendationCard key={rec.item_id} rec={rec} rank={idx + 1} showEvents userId="u_ai_reader" />
          ))}
        </div>
      )}
    </div>
  );
}
