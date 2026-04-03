'use client';

type BetEntry = {
  username: string;
  guess: number;
  wager: number;
};

type Props = {
  bets: BetEntry[];
  unit: string;
};

export default function BetTicker({ bets, unit }: Props) {
  if (bets.length === 0) {
    return <p className="text-gray-500 text-sm">No bets placed yet.</p>;
  }

  return (
    <div className="space-y-2 max-h-64 overflow-y-auto pr-1">
      {[...bets].reverse().map((b, i) => (
        <div
          key={i}
          className="flex items-center justify-between bg-gray-800 rounded-lg px-4 py-2 text-sm"
        >
          <span className="font-medium text-white w-28 truncate">{b.username}</span>
          <span className="text-violet-300 font-mono">
            {b.guess} <span className="text-gray-500 text-xs">{unit}</span>
          </span>
          <span className="text-green-400 font-mono">{b.wager} cr</span>
        </div>
      ))}
    </div>
  );
}
