'use client';

import { useState, useCallback } from 'react';
import { api, Recommendation } from '@/lib/api';
import EventLog, { LogEntry } from '@/components/EventLog';
import RecommendationCard from '@/components/RecommendationCard';

const EVENT_TYPES = [
  { type: 'view',            weight: 1, color: 'bg-gray-200 text-gray-700 hover:bg-gray-300' },
  { type: 'click',           weight: 2, color: 'bg-blue-100 text-blue-700 hover:bg-blue-200' },
  { type: 'read',            weight: 3, color: 'bg-purple-100 text-purple-700 hover:bg-purple-200' },
  { type: 'like',            weight: 5, color: 'bg-red-100 text-red-700 hover:bg-red-200' },
  { type: 'bookmark',        weight: 4, color: 'bg-yellow-100 text-yellow-700 hover:bg-yellow-200' },
  { type: 'share',           weight: 4, color: 'bg-green-100 text-green-700 hover:bg-green-200' },
  { type: 'purchase',        weight: 8, color: 'bg-emerald-100 text-emerald-700 hover:bg-emerald-200' },
  { type: 'inquiry',         weight: 6, color: 'bg-teal-100 text-teal-700 hover:bg-teal-200' },
  { type: 'meeting_request', weight: 7, color: 'bg-indigo-100 text-indigo-700 hover:bg-indigo-200' },
  { type: 'hide',            weight: -2, color: 'bg-orange-100 text-orange-700 hover:bg-orange-200' },
];

export default function SimulatorPage() {
  const [userId, setUserId] = useState('u_ai_reader');
  const [itemId, setItemId] = useState('news_001');
  const [log, setLog] = useState<LogEntry[]>([]);
  const [recs, setRecs] = useState<Recommendation[]>([]);
  const [recsLoading, setRecsLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const submitEvent = useCallback(async (feedbackType: string) => {
    if (!userId || !itemId) return;
    setSubmitting(true);
    try {
      const result = await api.postEvent(feedbackType, userId, itemId);
      const entry: LogEntry = {
        id: Date.now().toString(),
        feedbackType,
        userId,
        itemId,
        timestamp: new Date().toISOString(),
        success: result.success,
      };
      setLog((prev) => [entry, ...prev].slice(0, 50));
    } catch {
      const entry: LogEntry = {
        id: Date.now().toString(),
        feedbackType,
        userId,
        itemId,
        timestamp: new Date().toISOString(),
        success: false,
      };
      setLog((prev) => [entry, ...prev].slice(0, 50));
    } finally {
      setSubmitting(false);
    }
  }, [userId, itemId]);

  const refreshRecs = useCallback(async () => {
    setRecsLoading(true);
    try {
      // Try to detect category from itemId prefix
      if (itemId.startsWith('news_')) {
        const d = await api.getPersonalizedNews(userId);
        setRecs(d.recommendations || []);
      } else if (itemId.startsWith('prod_')) {
        const d = await api.getProductRecommendations(userId);
        setRecs(d.recommendations || []);
      } else if (itemId.startsWith('booth_')) {
        const d = await api.getExpoBooths(userId);
        setRecs(d.recommendations || []);
      } else {
        const d = await api.getPersonalizedNews(userId);
        setRecs(d.recommendations || []);
      }
    } catch {
      // ignore
    } finally {
      setRecsLoading(false);
    }
  }, [userId, itemId]);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Event Simulator</h1>
        <p className="text-sm text-gray-500 mt-1">Simulate user interactions to see how recommendations change in real-time</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Left: Simulator controls */}
        <div className="space-y-5">
          <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
            <h2 className="font-semibold text-gray-800 mb-4">Interaction Controls</h2>

            <div className="space-y-3 mb-4">
              <div>
                <label className="text-xs font-medium text-gray-600 block mb-1">User ID</label>
                <input
                  type="text"
                  value={userId}
                  onChange={(e) => setUserId(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                  placeholder="e.g. u_ai_reader"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-gray-600 block mb-1">Item ID</label>
                <input
                  type="text"
                  value={itemId}
                  onChange={(e) => setItemId(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                  placeholder="e.g. news_001, prod_001, booth_001"
                />
              </div>
            </div>

            <p className="text-xs font-medium text-gray-600 mb-2">Event Type (click to send):</p>
            <div className="grid grid-cols-2 gap-2">
              {EVENT_TYPES.map(({ type, weight, color }) => (
                <button
                  key={type}
                  onClick={() => submitEvent(type)}
                  disabled={submitting}
                  className={`px-3 py-2 rounded-lg text-xs font-semibold capitalize transition-colors disabled:opacity-50 flex items-center justify-between ${color}`}
                >
                  <span>{type.replace('_', ' ')}</span>
                  <span className="font-mono text-xs opacity-70">w={weight}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
            <div className="flex items-center justify-between mb-3">
              <h2 className="font-semibold text-gray-800">Event Log</h2>
              <button
                onClick={() => setLog([])}
                className="text-xs text-gray-400 hover:text-gray-600"
              >
                Clear
              </button>
            </div>
            <EventLog entries={log} />
          </div>
        </div>

        {/* Right: Recommendations */}
        <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-gray-800">Recommendations for {userId}</h2>
            <button
              onClick={refreshRecs}
              disabled={recsLoading}
              className="px-3 py-1.5 bg-brand-600 text-white rounded-lg text-xs font-medium hover:bg-brand-700 disabled:opacity-50"
            >
              {recsLoading ? 'Refreshing...' : 'Refresh Recs'}
            </button>
          </div>

          {recsLoading ? (
            <div className="space-y-2">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="rounded-xl border h-16 animate-pulse bg-gray-50" />
              ))}
            </div>
          ) : recs.length > 0 ? (
            <div className="space-y-2 max-h-[600px] overflow-y-auto pr-1">
              {recs.map((r, i) => (
                <RecommendationCard key={r.item_id} rec={r} rank={i + 1} />
              ))}
            </div>
          ) : (
            <div className="text-center py-12 text-gray-400 text-sm">
              Click Refresh Recs after submitting events to see updated recommendations.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
