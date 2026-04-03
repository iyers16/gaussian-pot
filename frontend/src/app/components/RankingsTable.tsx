'use client';

type RankEntry = {
  rank: number;
  username: string;
  guess: number;
  wager: number;
  payout: number;
  net: number;
};

type Props = {
  rankings: RankEntry[];
  unit: string;
  targetValue: number;
};

export default function RankingsTable({ rankings, unit, targetValue }: Props) {
  if (rankings.length === 0) return null;

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-semibold text-gray-300 uppercase tracking-wider">Final Rankings</h3>
        <span className="text-xs text-gray-500">
          Target: <span className="text-red-400 font-mono">{targetValue} {unit}</span>
        </span>
      </div>
      <div className="overflow-x-auto rounded-lg border border-gray-700">
        <table className="w-full text-sm">
          <thead className="bg-gray-800 text-gray-400 text-xs uppercase tracking-wider">
            <tr>
              <th className="px-3 py-2 text-left">Rank</th>
              <th className="px-3 py-2 text-left">Player</th>
              <th className="px-3 py-2 text-right">Guess</th>
              <th className="px-3 py-2 text-right">Wager</th>
              <th className="px-3 py-2 text-right">Payout</th>
              <th className="px-3 py-2 text-right">Net</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {rankings.map(r => (
              <tr key={r.rank} className="bg-gray-900 hover:bg-gray-800 transition-colors">
                <td className="px-3 py-2 font-bold text-gray-400">
                  {r.rank === 1 ? '🥇' : r.rank === 2 ? '🥈' : r.rank === 3 ? '🥉' : `#${r.rank}`}
                </td>
                <td className="px-3 py-2 text-white font-medium">{r.username}</td>
                <td className="px-3 py-2 text-right font-mono text-violet-300">{r.guess} <span className="text-gray-500 text-xs">{unit}</span></td>
                <td className="px-3 py-2 text-right font-mono text-gray-300">{r.wager.toFixed(2)}</td>
                <td className="px-3 py-2 text-right font-mono text-green-400">{r.payout.toFixed(2)}</td>
                <td className={`px-3 py-2 text-right font-mono font-semibold ${r.net >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                  {r.net >= 0 ? '+' : ''}{r.net.toFixed(2)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
