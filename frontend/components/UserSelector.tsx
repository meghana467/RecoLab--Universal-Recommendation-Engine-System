'use client';

interface Props {
  value: string;
  onChange: (val: string) => void;
  users?: string[];
  label?: string;
}

const DEFAULT_USERS = [
  'u_ai_reader',
  'u_trade_reader',
  'u_sports_reader',
  'buyer_energy_001',
  'buyer_textile_001',
  'buyer_energy_009',
  'buyer_import_014',
  ...Array.from({ length: 10 }, (_, i) => `u_${String(i + 1).padStart(3, '0')}`),
];

export default function UserSelector({ value, onChange, users = DEFAULT_USERS, label = 'Select User' }: Props) {
  return (
    <div className="flex items-center gap-2">
      <label className="text-sm font-medium text-gray-700">{label}:</label>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="border border-gray-300 rounded-lg px-3 py-1.5 text-sm bg-white focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent"
      >
        {users.map((u) => (
          <option key={u} value={u}>
            {u}
          </option>
        ))}
      </select>
    </div>
  );
}
