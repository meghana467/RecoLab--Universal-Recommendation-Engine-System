'use client';

import { useState, useEffect, useCallback } from 'react';
import { api, Recommendation, Buyer, BuyerResponse } from '@/lib/api';

// ── constants ──────────────────────────────────────────────────────────────

const DEMO_USERS = [
  { id: 'u_ai_reader',     label: 'AI Reader',     emoji: '🤖', color: 'purple', interests: ['ai','technology','startup','semiconductor'] },
  { id: 'u_trade_reader',  label: 'Trade Reader',  emoji: '📦', color: 'teal',   interests: ['trade','export','manufacturing','b2b'] },
  { id: 'u_sports_reader', label: 'Sports Reader', emoji: '⚽', color: 'amber',  interests: ['sports','entertainment','lifestyle'] },
];

const BOOTHS = [
  { id: 'booth_001', name: 'SunRise Solar – Photovoltaic Modules' },
  { id: 'booth_002', name: 'GreenWatt Energy Solutions' },
  { id: 'booth_011', name: 'Bharat Textiles Export House' },
  { id: 'booth_021', name: 'ChemIndia Specialties' },
  { id: 'booth_031', name: 'IndiaFoods Global Exports' },
  { id: 'booth_041', name: 'TechNova Electronics' },
];

type TabId = 'booths' | 'suppliers' | 'sessions' | 'trending' | 'seller';

const TABS: { id: TabId; label: string; icon: string; desc: string }[] = [
  { id: 'booths',    label: 'Buyer → Booth',     icon: '🏢', desc: 'Booths matched to buyer interests' },
  { id: 'suppliers', label: 'Buyer → Supplier',  icon: '🤝', desc: 'Suppliers sourced for each buyer' },
  { id: 'sessions',  label: 'Recommended Sessions', icon: '🎤', desc: 'Sessions curated per buyer profile' },
  { id: 'trending',  label: 'Trending Booths',   icon: '🔥', desc: 'Most-visited booths right now' },
  { id: 'seller',    label: 'Seller → Buyer',    icon: '👥', desc: 'Buyers matched to your booth' },
];

// ── label colour map ────────────────────────────────────────────────────────

const TAG_COLORS: Record<string, string> = {
  ai: 'bg-purple-100 text-purple-700',
  technology: 'bg-blue-100 text-blue-700',
  startup: 'bg-green-100 text-green-700',
  trade: 'bg-teal-100 text-teal-700',
  export: 'bg-cyan-100 text-cyan-700',
  manufacturing: 'bg-orange-100 text-orange-700',
  b2b: 'bg-indigo-100 text-indigo-700',
  solar: 'bg-amber-100 text-amber-700',
  textile: 'bg-pink-100 text-pink-700',
  chemicals: 'bg-red-100 text-red-700',
  food: 'bg-lime-100 text-lime-700',
  sports: 'bg-yellow-100 text-yellow-700',
  entertainment: 'bg-rose-100 text-rose-700',
  semiconductor: 'bg-violet-100 text-violet-700',
};
function tagClass(l: string) { return TAG_COLORS[l.toLowerCase()] ?? 'bg-gray-100 text-gray-600'; }

// ── colour helpers ──────────────────────────────────────────────────────────

const USER_COLORS: Record<string, { header: string; badge: string; ring: string; bar: string; num: string }> = {
  purple: { header: 'bg-purple-50 border-purple-200', badge: 'bg-purple-100 text-purple-800', ring: 'ring-purple-400', bar: 'bg-purple-500', num: 'bg-purple-600' },
  teal:   { header: 'bg-teal-50 border-teal-200',     badge: 'bg-teal-100 text-teal-800',     ring: 'ring-teal-400',   bar: 'bg-teal-500',   num: 'bg-teal-600'   },
  amber:  { header: 'bg-amber-50 border-amber-200',   badge: 'bg-amber-100 text-amber-800',   ring: 'ring-amber-400',  bar: 'bg-amber-500',  num: 'bg-amber-600'  },
};

// ── small components ────────────────────────────────────────────────────────

function Tag({ label }: { label: string }) {
  return <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${tagClass(label)}`}>{label}</span>;
}

function ScoreBar({ score, color }: { score: number; color: string }) {
  const pct = Math.min(Math.round(score * 100), 100);
  return (
    <div className="flex items-center gap-2 mt-2">
      <div className="flex-1 bg-gray-100 rounded-full h-1.5">
        <div className={`${color} h-1.5 rounded-full transition-all`} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-xs font-semibold text-gray-600 w-9 text-right">{score.toFixed(2)}</span>
    </div>
  );
}

function BoothCard({ rec, rank, barColor, numColor }: { rec: Recommendation; rank: number; barColor: string; numColor: string }) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-3 shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-start gap-2">
        <span className={`flex-shrink-0 w-6 h-6 rounded-full ${numColor} text-white text-xs font-bold flex items-center justify-center`}>{rank}</span>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-gray-900 line-clamp-2 leading-snug">{rec.title || rec.item_id}</p>
          <p className="text-xs text-gray-500 italic mt-0.5 line-clamp-1">{rec.reason}</p>
          <ScoreBar score={rec.score} color={barColor} />
          {rec.labels && rec.labels.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {rec.labels.slice(0, 4).map(l => <Tag key={l} label={l} />)}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function SessionCard({ rec, rank }: { rec: Recommendation; rank: number }) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-3 shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-start gap-2">
        <span className="flex-shrink-0 w-6 h-6 rounded-full bg-indigo-600 text-white text-xs font-bold flex items-center justify-center">{rank}</span>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-gray-900 leading-snug line-clamp-2">{rec.title || rec.item_id}</p>
          <p className="text-xs text-indigo-600 italic mt-0.5">{rec.reason}</p>
          <ScoreBar score={rec.score} color="bg-indigo-500" />
          {rec.labels && rec.labels.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {rec.labels.slice(0, 4).map(l => <Tag key={l} label={l} />)}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function BuyerCard({ buyer, rank }: { buyer: Buyer; rank: number }) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-3 shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-start gap-2">
        <span className="flex-shrink-0 w-6 h-6 rounded-full bg-rose-600 text-white text-xs font-bold flex items-center justify-center">{rank}</span>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-gray-900">{buyer.name || buyer.user_id}</p>
          <p className="text-xs text-gray-400 font-mono">{buyer.user_id}</p>
          <p className="text-xs text-gray-500 italic mt-0.5">{buyer.reason}</p>
          <ScoreBar score={buyer.score} color="bg-rose-500" />
          {buyer.labels && buyer.labels.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {buyer.labels.slice(0, 4).map(l => <Tag key={l} label={l} />)}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function EmptyState({ message }: { message: string }) {
  return (
    <div className="text-center py-10 text-gray-400 text-sm flex flex-col items-center gap-2">
      <span className="text-3xl">🔍</span>
      <span>{message}</span>
    </div>
  );
}

function LoadingCards() {
  return (
    <div className="space-y-2">
      {[...Array(5)].map((_, i) => (
        <div key={i} className="bg-gray-100 rounded-xl h-16 animate-pulse" />
      ))}
    </div>
  );
}

// ── buyer column (used in booths, suppliers, sessions tabs) ─────────────────

type FetchFn = (userId: string) => Promise<{ recommendations: Recommendation[]; user_interests?: string[] }>;

function BuyerColumn({ user, fetchFn, cardType }: {
  user: typeof DEMO_USERS[0];
  fetchFn: FetchFn;
  cardType: 'booth' | 'session';
}) {
  const [recs, setRecs] = useState<Recommendation[]>([]);
  const [interests, setInterests] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const c = USER_COLORS[user.color];

  async function load() {
    setLoading(true); setError(null);
    try {
      const d = await fetchFn(user.id);
      setRecs(d.recommendations || []);
      setInterests(d.user_interests || []);
    } catch (e) { setError(e instanceof Error ? e.message : 'Error'); }
    finally { setLoading(false); }
  }

  return (
    <div className={`flex-1 min-w-0 bg-white rounded-xl border ${c.header} border overflow-hidden shadow-sm`}>
      <div className={`px-4 py-3 border-b ${c.header}`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-xl">{user.emoji}</span>
            <div>
              <p className="text-sm font-semibold">{user.label}</p>
              <p className="text-xs font-mono opacity-60">{user.id}</p>
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
            {interests.map(l => <span key={l} className={`text-xs ${c.badge} px-2 py-0.5 rounded-full font-medium`}>{l}</span>)}
          </div>
        )}
      </div>
      <div className="p-3 space-y-2 min-h-[200px]">
        {error && <p className="text-xs text-red-500">{error}</p>}
        {loading ? <LoadingCards /> : recs.length > 0 ? (
          recs.map((r, i) =>
            cardType === 'booth'
              ? <BoothCard key={r.item_id} rec={r} rank={i + 1} barColor={c.bar} numColor={c.num} />
              : <SessionCard key={r.item_id} rec={r} rank={i + 1} />
          )
        ) : (
          <EmptyState message="Click Load to get recommendations" />
        )}
      </div>
    </div>
  );
}

// ── trending tab ─────────────────────────────────────────────────────────────

function TrendingTab() {
  const [recs, setRecs] = useState<Recommendation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const d = await api.getTrendingBooths(12);
      setRecs(d.recommendations || []);
    } catch (e) { setError(e instanceof Error ? e.message : 'Error'); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { load(); }, [load]);

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <p className="text-sm text-gray-500">Most-visited booths based on real-time interest signals</p>
        <button onClick={load} className="px-3 py-1.5 bg-orange-600 text-white rounded-lg text-xs font-medium hover:bg-orange-700">Refresh</button>
      </div>
      {error && <p className="text-sm text-red-500 mb-3">{error}</p>}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {[...Array(8)].map((_, i) => <div key={i} className="bg-gray-100 rounded-xl h-20 animate-pulse" />)}
        </div>
      ) : recs.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {recs.map((rec, i) => (
            <div key={rec.item_id} className="bg-white rounded-xl border border-orange-100 p-3 shadow-sm hover:shadow-md transition-shadow">
              <div className="flex items-start gap-3">
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-orange-500 text-white text-sm font-bold flex items-center justify-center">
                  {i + 1}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-gray-900 line-clamp-2">{rec.title}</p>
                  <p className="text-xs text-orange-600 italic mt-0.5">{rec.reason}</p>
                  <div className="mt-2 flex items-center gap-2">
                    <div className="flex-1 bg-gray-100 rounded-full h-1.5">
                      <div className="bg-orange-500 h-1.5 rounded-full" style={{ width: `${Math.min(rec.score * 100, 100)}%` }} />
                    </div>
                    <span className="text-xs font-semibold text-orange-600 w-9 text-right">{rec.score.toFixed(2)}</span>
                  </div>
                  {rec.labels && rec.labels.length > 0 && (
                    <div className="mt-2 flex flex-wrap gap-1">
                      {rec.labels.slice(0, 4).map(l => <Tag key={l} label={l} />)}
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <EmptyState message="No trending booths. Seed demo data from the Dashboard." />
      )}
    </div>
  );
}

// ── seller tab ───────────────────────────────────────────────────────────────

function SellerTab() {
  const [sellerId, setSellerId] = useState(BOOTHS[0].id);
  const [result, setResult] = useState<BuyerResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function load() {
    setLoading(true); setError(null);
    try {
      const d = await api.getSellerBuyers(sellerId);
      setResult(d);
    } catch (e) { setError(e instanceof Error ? e.message : 'Error'); }
    finally { setLoading(false); }
  }

  return (
    <div>
      {/* Seller selector */}
      <div className="bg-rose-50 border border-rose-200 rounded-xl p-4 mb-5">
        <p className="text-xs font-semibold text-rose-700 uppercase tracking-wide mb-2">Select Your Booth / Company</p>
        <div className="flex flex-wrap gap-2">
          {BOOTHS.map(b => (
            <button
              key={b.id}
              onClick={() => setSellerId(b.id)}
              className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-all ${
                sellerId === b.id
                  ? 'bg-rose-600 text-white border-rose-600 shadow-sm'
                  : 'bg-white text-gray-700 border-gray-300 hover:border-rose-400 hover:text-rose-700'
              }`}
            >
              {b.name}
            </button>
          ))}
        </div>
        <button
          onClick={load}
          disabled={loading}
          className="mt-3 px-4 py-2 bg-rose-600 text-white rounded-lg text-sm font-medium hover:bg-rose-700 disabled:opacity-50 transition"
        >
          {loading ? 'Finding buyers...' : '👥 Find Matching Buyers'}
        </button>
      </div>

      {/* Booth info */}
      {result && (
        <div className="mb-4 flex items-center gap-2 flex-wrap">
          <span className="text-sm font-semibold text-gray-700">{result.seller_name}</span>
          {result.seller_labels?.map(l => <Tag key={l} label={l} />)}
          <span className="ml-auto text-xs text-gray-400">{result.count} potential buyers found</span>
        </div>
      )}

      {error && <p className="text-sm text-red-500 mb-3">{error}</p>}

      {loading ? (
        <div className="space-y-2">{[...Array(5)].map((_, i) => <div key={i} className="bg-gray-100 rounded-xl h-16 animate-pulse" />)}</div>
      ) : result && result.buyers.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {result.buyers.map((b, i) => <BuyerCard key={b.user_id} buyer={b} rank={i + 1} />)}
        </div>
      ) : result ? (
        <EmptyState message="No matching buyers found. Seed demo data and try again." />
      ) : (
        <EmptyState message="Select a booth and click Find Matching Buyers" />
      )}
    </div>
  );
}

// ── main page ─────────────────────────────────────────────────────────────────

export default function ExpoPage() {
  const [tab, setTab] = useState<TabId>('booths');

  return (
    <div>
      {/* Page header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Expo & Trade Recommendations</h1>
        <p className="text-sm text-gray-500 mt-1">
          AI-powered matchmaking for trade shows, B2B expos, and supplier discovery
        </p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 bg-gray-100 p-1 rounded-xl mb-6 overflow-x-auto">
        {TABS.map(t => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-all ${
              tab === t.id
                ? 'bg-white shadow-sm text-gray-900'
                : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            <span>{t.icon}</span>
            <span>{t.label}</span>
          </button>
        ))}
      </div>

      {/* Tab description */}
      <p className="text-xs text-gray-400 mb-4 italic">
        {TABS.find(t => t.id === tab)?.desc}
      </p>

      {/* Tab content */}
      {tab === 'booths' && (
        <div className="flex gap-4 overflow-x-auto pb-2">
          {DEMO_USERS.map(u => (
            <BuyerColumn
              key={u.id}
              user={u}
              fetchFn={(id) => api.getExpoBooths(id)}
              cardType="booth"
            />
          ))}
        </div>
      )}

      {tab === 'suppliers' && (
        <div className="flex gap-4 overflow-x-auto pb-2">
          {DEMO_USERS.map(u => (
            <BuyerColumn
              key={u.id}
              user={u}
              fetchFn={(id) => api.getExpoSuppliers(id)}
              cardType="booth"
            />
          ))}
        </div>
      )}

      {tab === 'sessions' && (
        <div className="flex gap-4 overflow-x-auto pb-2">
          {DEMO_USERS.map(u => (
            <BuyerColumn
              key={u.id}
              user={u}
              fetchFn={(id) => api.getExpoSessions(id)}
              cardType="session"
            />
          ))}
        </div>
      )}

      {tab === 'trending' && <TrendingTab />}

      {tab === 'seller' && <SellerTab />}
    </div>
  );
}
