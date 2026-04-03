'use client';

type CurvePoint = { x: number; y: number };
type BetOnLine = { username: string; guess: number; wager: number };

type Props = {
  curvePoints: CurvePoint[];
  betsOnNumberline: BetOnLine[];
  targetValue?: number | null;
  unit: string;
  showTarget: boolean;
};

export default function GaussianCurve({ curvePoints, betsOnNumberline, targetValue, unit, showTarget }: Props) {
  if (curvePoints.length === 0) {
    return <div className="h-40 flex items-center justify-center text-gray-500 text-sm">No data yet</div>;
  }

  const W = 600;
  const H = 160;
  const PAD = { top: 10, bottom: 30, left: 40, right: 20 };
  const chartW = W - PAD.left - PAD.right;
  const chartH = H - PAD.top - PAD.bottom;

  const xs = curvePoints.map(p => p.x);
  const ys = curvePoints.map(p => p.y);
  const xMin = Math.min(...xs);
  const xMax = Math.max(...xs);
  const yMax = Math.max(...ys, 0.01);
  const xRange = xMax - xMin || 1;

  const toSvgX = (x: number) => PAD.left + ((x - xMin) / xRange) * chartW;
  const toSvgY = (y: number) => PAD.top + chartH - (y / yMax) * chartH;

  const pathD = curvePoints.reduce((acc, p, i) => {
    const sx = toSvgX(p.x);
    const sy = toSvgY(p.y);
    return acc + (i === 0 ? `M ${sx} ${sy}` : ` L ${sx} ${sy}`);
  }, '');

  // Fill area under curve.
  const firstX = toSvgX(curvePoints[0].x);
  const lastX = toSvgX(curvePoints[curvePoints.length - 1].x);
  const baseline = PAD.top + chartH;
  const fillD = `${pathD} L ${lastX} ${baseline} L ${firstX} ${baseline} Z`;

  return (
    <div className="w-full overflow-x-auto">
      <svg viewBox={`0 0 ${W} ${H}`} className="w-full" style={{ minWidth: 280 }}>
        {/* Grid lines */}
        {[0.25, 0.5, 0.75, 1].map(frac => (
          <line
            key={frac}
            x1={PAD.left} y1={toSvgY(frac * yMax)}
            x2={PAD.left + chartW} y2={toSvgY(frac * yMax)}
            stroke="#374151" strokeWidth={0.5}
          />
        ))}

        {/* Fill */}
        <path d={fillD} fill="rgba(139,92,246,0.15)" />

        {/* Curve */}
        <path d={pathD} fill="none" stroke="#7c3aed" strokeWidth={2} />

        {/* Target line (host view) */}
        {showTarget && targetValue != null && (
          <g>
            <line
              x1={toSvgX(targetValue)} y1={PAD.top}
              x2={toSvgX(targetValue)} y2={baseline}
              stroke="#ef4444" strokeWidth={1.5} strokeDasharray="4 2"
            />
            <text
              x={toSvgX(targetValue) + 4} y={PAD.top + 12}
              fill="#ef4444" fontSize={10} fontFamily="monospace"
            >
              {targetValue} {unit}
            </text>
          </g>
        )}

        {/* Bet markers */}
        {betsOnNumberline.map((b, i) => (
          <g key={i}>
            <line
              x1={toSvgX(b.guess)} y1={baseline - 6}
              x2={toSvgX(b.guess)} y2={baseline + 2}
              stroke="#a78bfa" strokeWidth={1.5}
            />
            <text
              x={toSvgX(b.guess)} y={baseline + 12}
              textAnchor="middle" fill="#a78bfa" fontSize={8} fontFamily="monospace"
            >
              {b.username.substring(0, 5)}
            </text>
          </g>
        ))}

        {/* X-axis */}
        <line x1={PAD.left} y1={baseline} x2={PAD.left + chartW} y2={baseline} stroke="#4b5563" strokeWidth={1} />
        <text x={PAD.left} y={H - 2} fill="#6b7280" fontSize={9} fontFamily="monospace">{xMin.toFixed(1)}</text>
        <text x={PAD.left + chartW} y={H - 2} textAnchor="end" fill="#6b7280" fontSize={9} fontFamily="monospace">{xMax.toFixed(1)} {unit}</text>
      </svg>
    </div>
  );
}
