'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { api, InterestOption, OnboardResponse } from '@/lib/api';

const CATEGORY_LABELS: Record<string, string> = {
  news: '📰 News & Media',
  products: '🛍️ Shopping',
  expo: '🏭 Expo & Trade',
};

export default function OnboardPage() {
  const router = useRouter();
  const [userId, setUserId] = useState('new_user_' + Math.random().toString(36).slice(2, 8));
  const [options, setOptions] = useState<InterestOption[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<OnboardResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getInterestOptions()
      .then(d => setOptions(d.interests))
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  function toggle(label: string) {
    setSelected(prev => {
      const next = new Set(prev);
      if (next.has(label)) next.delete(label);
      else if (next.size < 8) next.add(label);
      return next;
    });
  }

  async function submit() {
    if (selected.size < 3) return;
    setSubmitting(true);
    setError(null);
    try {
      const data = await api.postOnboard(userId, Array.from(selected));
      setResult(data);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Error');
    } finally {
      setSubmitting(false);
    }
  }

  // Group options by category
  const byCategory = options.reduce<Record<string, InterestOption[]>>((acc, opt) => {
    if (!acc[opt.category]) acc[opt.category] = [];
    acc[opt.category].push(opt);
    return acc;
  }, {});

  if (result) {
    return (
      <div className="max-w-3xl mx-auto">
        {/* Success banner */}
        <div className="bg-green-50 border border-green-200 rounded-2xl p-5 mb-6">
          <div className="flex items-center gap-3 mb-3">
            <span className="text-2xl">🎉</span>
            <div>
              <p className="font-bold text-green-900">Welcome, {result.user_id}!</p>
              <p className="text-sm text-green-700">{result.message}</p>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {result.interests.map(l => (
              <span key={l} className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded-full font-medium">
                {l}
              </span>
            ))}
          </div>
        </div>

        {/* Immediate recommendations across all scenarios */}
        <div className="space-y-6">
          {[
            { key: 'news', label: '📰 News for you', items: result.news },
            { key: 'products', label: '🛍️ Products for you', items: result.products },
            { key: 'expo', label: '🏭 Expo booths for you', items: result.expo },
          ].map(({ key, label, items }) => (
            <div key={key} className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
              <div className="px-4 py-3 border-b border-gray-100 bg-gray-50">
                <p className="font-semibold text-gray-900 text-sm">{label}</p>
                <p className="text-xs text-gray-400 mt-0.5">
                  Label-matched recommendations — personalised from day one
                </p>
              </div>
              <div className="p-3 space-y-2">
                {items.slice(0, 5).map((rec, i) => (
                  <div key={rec.item_id} className="flex items-start gap-2 p-2 rounded-lg hover:bg-gray-50">
                    <span className="w-5 h-5 rounded-full bg-brand-600 text-white text-xs font-bold flex items-center justify-center flex-shrink-0 mt-0.5">
                      {i + 1}
                    </span>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 line-clamp-1">{rec.title || rec.item_id}</p>
                      <p className="text-xs text-gray-500 italic">{rec.reason}</p>
                      {rec.labels?.length > 0 && (
                        <div className="mt-1 flex flex-wrap gap-1">
                          {rec.labels.slice(0, 4).map(l => (
                            <span key={l} className="text-xs bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded-full">
                              {l}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className="mt-6 flex gap-3">
          <button
            onClick={() => router.push('/personalized')}
            className="px-4 py-2 bg-brand-600 text-white rounded-lg text-sm font-medium hover:bg-brand-700"
          >
            Go to Personalized Feed →
          </button>
          <button
            onClick={() => { setResult(null); setSelected(new Set()); }}
            className="px-4 py-2 bg-white border border-gray-300 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-50"
          >
            Try another user
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">👋 New User Onboarding</h1>
        <p className="text-sm text-gray-500 mt-1">
          Cold-start solution — pick interests to get personalised recommendations immediately,
          before the collaborative filtering model has any data on you.
        </p>
      </div>

      {/* User ID input */}
      <div className="bg-blue-50 border border-blue-200 rounded-xl p-4 mb-5">
        <label className="block text-xs font-semibold text-blue-700 mb-1">User ID (simulated new user)</label>
        <input
          value={userId}
          onChange={e => setUserId(e.target.value)}
          className="w-full border border-blue-200 rounded-lg px-3 py-2 text-sm bg-white focus:outline-none focus:ring-2 focus:ring-blue-400"
          placeholder="e.g. new_user_abc123"
        />
      </div>

      {/* Interest picker */}
      {loading ? (
        <div className="space-y-4">
          {[1, 2, 3].map(i => <div key={i} className="h-24 bg-gray-100 rounded-xl animate-pulse" />)}
        </div>
      ) : (
        <div className="space-y-4">
          {Object.entries(byCategory).map(([cat, opts]) => (
            <div key={cat} className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
              <div className="px-4 py-2 bg-gray-50 border-b border-gray-100">
                <p className="text-xs font-semibold text-gray-600 uppercase tracking-wide">
                  {CATEGORY_LABELS[cat] ?? cat}
                </p>
              </div>
              <div className="p-3 flex flex-wrap gap-2">
                {opts.map(opt => {
                  const active = selected.has(opt.label);
                  return (
                    <button
                      key={opt.label}
                      onClick={() => toggle(opt.label)}
                      className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium border transition-all ${
                        active
                          ? 'bg-brand-600 text-white border-brand-600 shadow-sm'
                          : 'bg-white text-gray-700 border-gray-300 hover:border-brand-400 hover:text-brand-700'
                      }`}
                    >
                      <span>{opt.emoji}</span>
                      <span className="capitalize">{opt.label}</span>
                    </button>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Selection counter + submit */}
      <div className="mt-5 flex items-center justify-between">
        <p className="text-sm text-gray-500">
          {selected.size} selected
          <span className="text-gray-400"> (pick 3–8)</span>
        </p>
        <button
          onClick={submit}
          disabled={selected.size < 3 || submitting}
          className="px-6 py-2.5 bg-brand-600 text-white rounded-xl text-sm font-semibold hover:bg-brand-700 disabled:opacity-40 disabled:cursor-not-allowed transition"
        >
          {submitting ? 'Saving…' : 'Get my recommendations →'}
        </button>
      </div>

      {error && <p className="mt-3 text-sm text-red-500">{error}</p>}

      {/* How it works */}
      <div className="mt-8 bg-gray-50 border border-gray-200 rounded-xl p-4">
        <p className="text-xs font-semibold text-gray-600 uppercase tracking-wide mb-2">How cold-start is solved here</p>
        <div className="space-y-2 text-xs text-gray-500">
          <div className="flex gap-2"><span className="font-semibold text-brand-600">1. Interest Picker</span><span>Your picks become labels on your Gorse user profile immediately.</span></div>
          <div className="flex gap-2"><span className="font-semibold text-brand-600">2. Stratified Fallback</span><span>Results are drawn proportionally from each interest topic — not just the first alphabetically.</span></div>
          <div className="flex gap-2"><span className="font-semibold text-brand-600">3. Cross-Scenario</span><span>Your news interests automatically seed expo and product recommendations too.</span></div>
          <div className="flex gap-2"><span className="font-semibold text-brand-600">4. Label Similarity</span><span>New items with no interactions get surfaced by comparing their labels to yours.</span></div>
        </div>
      </div>
    </div>
  );
}
