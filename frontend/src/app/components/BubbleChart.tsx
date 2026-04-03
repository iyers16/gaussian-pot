'use client';

type BetOnLine = {
  username: string;
  guess: number;
  wager: number;
};

type Props = {
  bets: BetOnLine[];
  targetValue?: number | null;
  unit: string;
};

export default function BubbleChart({ bets, targetValue, unit }: Props) {
  if (bets.length === 0) {
    return <p className="text-gray-500 text-sm">Waiting for bets…</p>;
  }

  const values = bets.map(b => b.guess);
  if (targetValue != null) values.push(targetValue);
  const minV = Math.min(...values);
  const maxV = Math.max(...values);
  const range = maxV - minV || 1;

  const toPercent = (v: number) => ((v - minV) / range) * 88 + 6; // 6–94%

  const maxWager = Math.max(...bets.map(b => b.wager), 1);
  const bubbleSize = (wager: number) => 14 + (wager / maxWager) * 22;

  return (
    <div className="relative h-20 w-full rounded-lg bg-gray-800 overflow-hidden">
      {/* Number line */}
      <div className="absolute top-1/2 left-0 right-0 h-px bg-gray-600" />

      {/* Scale labels */}
      <span className="absolute top-1 left-2 text-xs text-gray-500 font-mono">{minV} {unit}</span>
      <span className="absolute top-1 right-2 text-xs text-gray-500 font-mono">{maxV} {unit}</span>

      {/* Target marker */}
      {targetValue != null && (
        <div
          className="absolute top-0 bottom-0 w-0.5 bg-red-500"
          style={{ left: `${toPercent(targetValue)}%` }}
        >
          <span className="absolute -top-1 left-1 text-xs text-red-400 font-mono whitespace-nowrap">
            ▼ {targetValue}
          </span>
        </div>
      )}

      {/* Bet bubbles */}
      {bets.map((b, i) => {
        const size = bubbleSize(b.wager);
        return (
          <div
            key={i}
            title={`${b.username}: ${b.guess} ${unit} (${b.wager} cr)`}
            className="absolute rounded-full bg-violet-500/80 border border-violet-400 flex items-center justify-center cursor-default"
            style={{
              width: size,
              height: size,
              left: `calc(${toPercent(b.guess)}% - ${size / 2}px)`,
              top: `calc(50% - ${size / 2}px)`,
            }}
          >
            <span className="text-[9px] text-white font-bold truncate px-0.5">
              {b.username.charAt(0).toUpperCase()}
            </span>
          </div>
        );
      })}
    </div>
  );
}
