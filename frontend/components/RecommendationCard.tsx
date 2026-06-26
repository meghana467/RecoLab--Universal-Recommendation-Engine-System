'use client';

import { useState } from 'react';
import { Recommendation, api } from '@/lib/api';

interface Props {
  rec: Recommendation;
  rank?: number;
  showEvents?: boolean;
  userId?: string;
}

const labelColors: Record<string, string> = {
  ai: 'bg-purple-100 text-purple-700',
  technology: 'bg-blue-100 text-blue-700',
  startup: 'bg-green-100 text-green-700',
  india: 'bg-orange-100 text-orange-700',
  sports: 'bg-yellow-100 text-yellow-700',
  trade: 'bg-teal-100 text-teal-700',
  finance: 'bg-indigo-100 text-indigo-700',
  solar: 'bg-amber-100 text-amber-700',
  textile: 'bg-pink-100 text-pink-700',
  chemicals: 'bg-red-100 text-red-700',
  food: 'bg-lime-100 text-lime-700',
  health: 'bg-emerald-100 text-emerald-700',
  electronics: 'bg-cyan-100 text-cyan-700',
  fashion: 'bg-rose-100 text-rose-700',
};

function labelClass(label: string) {
  return labelColors[label.toLowerCase()] ?? 'bg-gray-100 text-gray-600';
}

const EVENT_BUTTONS = [
  { type: 'like',     emoji: '👍', label: 'Like' },
  { type: 'share',    emoji: '🔗', label: 'Share' },
  { type: 'bookmark', emoji: '🔖', label: 'Bookmark' },
];

export default function RecommendationCard({ rec, rank, showEvents, userId }: Props) {
  const [sent, setSent] = useState<Record<string, boolean>>({});
  const [sending, setSending] = useState<string | null>(null);

  async function sendEvent(feedbackType: string) {
    if (!userId || sending) return;
    setSending(feedbackType);
    try {
      await api.postEvent(feedbackType, userId, rec.item_id);
      setSent((prev) => ({ ...prev, [feedbackType]: true }));
    } catch {
      // silently fail — demo environment
    } finally {
      setSending(null);
    }
  }

  const scorePercent = Math.min(Math.round(rec.score * 100), 100);

  return (
    <div className="bg-white rounded-xl border border-gray-200 p-4 shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-start gap-3">
        {rank !== undefined && (
          <span className="flex-shrink-0 w-7 h-7 rounded-full bg-brand-600 text-white text-xs font-bold flex items-center justify-center">
            {rank}
          </span>
        )}
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-semibold text-gray-900 leading-snug line-clamp-2">
            {rec.title || rec.item_id}
          </h3>
          <p className="mt-1 text-xs text-gray-500 italic line-clamp-2">{rec.reason}</p>

          {/* Score bar */}
          <div className="mt-2 flex items-center gap-2">
            <div className="flex-1 bg-gray-100 rounded-full h-1.5">
              <div
                className="bg-brand-500 h-1.5 rounded-full transition-all"
                style={{ width: `${scorePercent}%` }}
              />
            </div>
            <span className="text-xs font-medium text-brand-600 w-10 text-right">
              {rec.score.toFixed(2)}
            </span>
          </div>

          {/* Labels */}
          {rec.labels && rec.labels.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {rec.labels.slice(0, 5).map((l) => (
                <span key={l} className={`text-xs px-2 py-0.5 rounded-full font-medium ${labelClass(l)}`}>
                  {l}
                </span>
              ))}
            </div>
          )}

          {/* Event buttons */}
          {showEvents && userId && (
            <div className="mt-3 flex gap-2">
              {EVENT_BUTTONS.map(({ type, emoji, label }) => (
                <button
                  key={type}
                  onClick={() => sendEvent(type)}
                  disabled={sending === type || sent[type]}
                  title={label}
                  className={`flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-medium border transition-all
                    ${sent[type]
                      ? 'bg-brand-50 border-brand-300 text-brand-700'
                      : 'bg-gray-50 border-gray-200 text-gray-600 hover:bg-gray-100 hover:border-gray-300'
                    } disabled:opacity-60`}
                >
                  <span>{emoji}</span>
                  <span>{sent[type] ? '✓' : label}</span>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
