'use client';

export interface LogEntry {
  id: string;
  feedbackType: string;
  userId: string;
  itemId: string;
  timestamp: string;
  success: boolean;
}

interface Props {
  entries: LogEntry[];
}

const typeColors: Record<string, string> = {
  view: 'bg-gray-100 text-gray-600',
  click: 'bg-blue-100 text-blue-700',
  read: 'bg-purple-100 text-purple-700',
  like: 'bg-red-100 text-red-600',
  bookmark: 'bg-yellow-100 text-yellow-700',
  share: 'bg-green-100 text-green-700',
  purchase: 'bg-emerald-100 text-emerald-700',
  inquiry: 'bg-teal-100 text-teal-700',
  meeting_request: 'bg-indigo-100 text-indigo-700',
  hide: 'bg-orange-100 text-orange-700',
};

export default function EventLog({ entries }: Props) {
  if (entries.length === 0) {
    return (
      <div className="text-center text-gray-400 py-8 text-sm">
        No events recorded yet. Submit events above.
      </div>
    );
  }

  return (
    <div className="space-y-2 max-h-72 overflow-y-auto pr-1">
      {entries.map((e) => (
        <div
          key={e.id}
          className={`flex items-center gap-3 rounded-lg border px-3 py-2 text-xs ${
            e.success ? 'border-green-200 bg-green-50' : 'border-red-200 bg-red-50'
          }`}
        >
          <span className={`rounded-full px-2 py-0.5 font-semibold capitalize ${typeColors[e.feedbackType] ?? 'bg-gray-100 text-gray-600'}`}>
            {e.feedbackType}
          </span>
          <span className="font-mono text-gray-700">{e.userId}</span>
          <span className="text-gray-400">→</span>
          <span className="font-mono text-gray-700">{e.itemId}</span>
          <span className="ml-auto text-gray-400 whitespace-nowrap">
            {new Date(e.timestamp).toLocaleTimeString()}
          </span>
          {e.success ? (
            <span className="text-green-500">✓</span>
          ) : (
            <span className="text-red-500">✗</span>
          )}
        </div>
      ))}
    </div>
  );
}
