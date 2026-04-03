'use client';

import { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import BetTicker from '../components/BetTicker';
import BubbleChart from '../components/BubbleChart';
import GaussianCurve from '../components/GaussianCurve';
import RankingsTable from '../components/RankingsTable';
import DebtChecklist from '../components/DebtChecklist';
import { useWebSocket } from '@/hooks/useWebSocket';

type GameState = 'IDLE' | 'ROUND_OPEN' | 'ROUND_CALLED' | 'SETTLING';

type RoundInfo = {
  question_text: string;
  mode: string;
  unit: string;
};

type BetEntry = { username: string; guess: number; wager: number };
type CurvePoint = { x: number; y: number };
type RankEntry = { rank: number; username: string; guess: number; wager: number; payout: number; net: number };
type Debt = { id: number; payer_username: string; payee_username: string; amount: number; payer_confirmed: boolean; payee_confirmed: boolean; status: string };

export default function PlayerPage() {
  const router = useRouter();
  const [token, setToken] = useState('');
  const [username, setUsername] = useState('');
  const [credits, setCredits] = useState(0);

  const [gameState, setGameState] = useState<GameState>('IDLE');
  const [round, setRound] = useState<RoundInfo | null>(null);
  const [bets, setBets] = useState<BetEntry[]>([]);
  const [curvePoints, setCurvePoints] = useState<CurvePoint[]>([]);
  const [betsOnLine, setBetsOnLine] = useState<BetEntry[]>([]);
  const [rankings, setRankings] = useState<RankEntry[]>([]);
  const [targetValue, setTargetValue] = useState<number | null>(null);
  const [debts, setDebts] = useState<Debt[]>([]);
  const [myBet, setMyBet] = useState<{ guess: number; wager: number } | null>(null);

  const [guess, setGuess] = useState('');
  const [wager, setWager] = useState('');
  const [betError, setBetError] = useState('');
  const [betLoading, setBetLoading] = useState(false);

  useEffect(() => {
    const t = localStorage.getItem('gp_token');
    const u = localStorage.getItem('gp_username');
    const r = localStorage.getItem('gp_role');
    const c = localStorage.getItem('gp_credits');
    if (!t || r !== 'player') { router.push('/'); return; }
    setToken(t);
    setUsername(u || '');
    setCredits(Number(c) || 0);

    // Fetch current round state.
    fetch('/api/round', { headers: { Authorization: `Bearer ${t}` } })
      .then(r => r.json())
      .then(data => {
        if (data.state && data.state !== 'IDLE') {
          setGameState(data.state as GameState);
          setRound({ question_text: data.question_text, mode: data.mode, unit: data.unit });
          if (data.target_value != null) setTargetValue(data.target_value);
        }
      });

    // Fetch my debts.
    fetch('/api/debts', { headers: { Authorization: `Bearer ${t}` } })
      .then(r => r.json())
      .then(data => { if (Array.isArray(data)) setDebts(data); });
  }, [router]);

  const handleEvent = useCallback((event: { type: string; payload: unknown }) => {
    const p = event.payload as Record<string, unknown>;
    switch (event.type) {
      case 'round_opened':
        setGameState('ROUND_OPEN');
        setRound({ question_text: p.question_text as string, mode: p.mode as string, unit: p.unit as string });
        setBets([]);
        setCurvePoints([]);
        setBetsOnLine([]);
        setRankings([]);
        setTargetValue(null);
        setMyBet(null);
        setBetError('');
        break;

      case 'bet_ticker':
        setBets((p.bets as BetEntry[]) || []);
        break;

      case 'distribution_update': {
        const dist = p as { curve_points: CurvePoint[]; bets_on_numberline: BetEntry[] };
        setCurvePoints(dist.curve_points || []);
        setBetsOnLine(dist.bets_on_numberline || []);
        break;
      }

      case 'round_called': {
        setGameState('SETTLING');
        setTargetValue(p.target_value as number);
        setRankings((p.rankings as RankEntry[]) || []);
        setDebts(prev => {
          const incoming = (p.debts as Debt[]) || [];
          // Merge: keep existing + add new.
          const existingIDs = new Set(prev.map(d => d.id));
          return [...prev, ...incoming.filter(d => !existingIDs.has(d.id))];
        });
        const dist = p.distribution as { curve_points: CurvePoint[]; bets_on_numberline: BetEntry[] };
        if (dist) {
          setCurvePoints(dist.curve_points || []);
          setBetsOnLine(dist.bets_on_numberline || []);
        }
        break;
      }

      case 'debt_update': {
        const updated = p as Debt;
        setDebts(prev => prev.map(d => d.id === updated.id ? updated : d));
        break;
      }

      case 'round_settled':
        setGameState('IDLE');
        break;

      case 'session_update':
        break;
    }
  }, []);

  useWebSocket(handleEvent);

  async function placeBet(e: React.FormEvent) {
    e.preventDefault();
    setBetLoading(true);
    setBetError('');

    const res = await fetch('/api/round/bet', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ guess: Number(guess), wager: Number(wager) }),
    });
    const data = await res.json();
    if (!res.ok) {
      setBetError(data.error || 'Failed to place bet');
    } else {
      setMyBet({ guess: Number(guess), wager: Number(wager) });
      setCredits(prev => prev - Number(wager));
    }
    setBetLoading(false);
  }

  function handleDebtUpdate(updated: Debt) {
    setDebts(prev => prev.map(d => d.id === updated.id ? updated : d));
  }

  async function logout() {
    await fetch('/api/auth/logout', { method: 'POST', headers: { Authorization: `Bearer ${token}` } });
    localStorage.clear();
    router.push('/');
  }

  const modeDesc: Record<string, string> = {
    sniper: 'High risk — big wagers tighten the curve, rewarding precise guesses heavily.',
    social: 'Spread the love — big wagers flatten the curve, distributing payouts more evenly.',
    volatile: 'Chaos mode — curve width is driven by how spread out everyone\'s guesses are.',
  };

  const pendingDebts = debts.filter(d => d.status === 'PENDING');

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="sticky top-0 z-10 bg-gray-900 border-b border-gray-800 px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="font-bold text-violet-400">Gaussian Pot</span>
          <span className="text-gray-500 text-sm">{username}</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm font-mono text-green-400">{credits.toFixed(2)} cr</span>
          {pendingDebts.length > 0 && (
            <span className="text-xs bg-red-700 text-white px-2 py-0.5 rounded-full">{pendingDebts.length} debt{pendingDebts.length > 1 ? 's' : ''}</span>
          )}
          <button onClick={logout} className="text-xs text-gray-400 hover:text-white transition-colors">Leave</button>
        </div>
      </header>

      <main className="max-w-2xl mx-auto p-4 space-y-6">

        {/* IDLE */}
        {gameState === 'IDLE' && (
          <div className="text-center py-16">
            <div className="text-5xl mb-4">⏳</div>
            <h2 className="text-xl text-gray-300">Waiting for the host to open a round…</h2>
            {pendingDebts.length > 0 && (
              <div className="mt-8">
                <h3 className="text-sm font-semibold text-gray-400 mb-3 uppercase tracking-wider">Unsettled Debts</h3>
                <DebtChecklist debts={debts} myUsername={username} token={token} onUpdate={handleDebtUpdate} />
              </div>
            )}
          </div>
        )}

        {/* ROUND_OPEN */}
        {gameState === 'ROUND_OPEN' && round && (
          <div className="space-y-5">
            <div className="bg-gray-900 rounded-xl p-5 border border-gray-800">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-xs uppercase tracking-wider text-violet-400 font-semibold">{round.mode} mode</span>
              </div>
              <h2 className="text-lg font-semibold text-white mb-2">{round.question_text}</h2>
              <p className="text-xs text-gray-500">{modeDesc[round.mode]}</p>
            </div>

            {!myBet ? (
              <form onSubmit={placeBet} className="bg-gray-900 rounded-xl p-5 border border-gray-800 space-y-4">
                <h3 className="text-sm font-semibold text-gray-300 uppercase tracking-wider">Place Your Bet</h3>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs text-gray-500 mb-1">Your Guess ({round.unit})</label>
                    <input
                      type="number"
                      value={guess}
                      onChange={e => setGuess(e.target.value)}
                      placeholder="0"
                      required
                      className="w-full bg-gray-800 text-white rounded-lg px-3 py-2 text-sm border border-gray-700 focus:border-violet-500 outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500 mb-1">Wager (cr, max {credits.toFixed(0)})</label>
                    <input
                      type="number"
                      value={wager}
                      onChange={e => setWager(e.target.value)}
                      placeholder="50"
                      min={1}
                      max={credits}
                      required
                      className="w-full bg-gray-800 text-white rounded-lg px-3 py-2 text-sm border border-gray-700 focus:border-violet-500 outline-none"
                    />
                  </div>
                </div>
                {betError && <p className="text-red-400 text-xs">{betError}</p>}
                <button
                  type="submit"
                  disabled={betLoading || !guess || !wager}
                  className="w-full bg-violet-600 hover:bg-violet-500 disabled:bg-gray-700 disabled:text-gray-500 text-white font-semibold rounded-lg py-2.5 text-sm transition-colors"
                >
                  {betLoading ? 'Placing…' : 'Lock In Bet'}
                </button>
              </form>
            ) : (
              <div className="bg-green-900/30 rounded-xl p-5 border border-green-800 text-center">
                <p className="text-green-400 font-medium">Bet placed!</p>
                <p className="text-gray-400 text-sm mt-1">
                  {myBet.guess} {round.unit} · {myBet.wager} cr wagered
                </p>
                <p className="text-gray-500 text-xs mt-1">Waiting for the host to call the round…</p>
              </div>
            )}

            <div className="bg-gray-900 rounded-xl p-5 border border-gray-800 space-y-4">
              <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Live Bets</h3>
              <BubbleChart bets={betsOnLine} unit={round.unit} />
              <GaussianCurve curvePoints={curvePoints} betsOnNumberline={betsOnLine} unit={round.unit} showTarget={false} />
              <BetTicker bets={bets} unit={round.unit} />
            </div>
          </div>
        )}

        {/* SETTLING / ROUND_CALLED */}
        {(gameState === 'SETTLING' || gameState === 'ROUND_CALLED') && (
          <div className="space-y-5">
            {round && targetValue != null && (
              <>
                <div className="bg-gray-900 rounded-xl p-5 border border-gray-800">
                  <p className="text-xs text-gray-500 mb-1">{round.question_text}</p>
                  <div className="flex items-baseline gap-2">
                    <span className="text-3xl font-bold font-mono text-red-400">{targetValue}</span>
                    <span className="text-gray-400">{round.unit}</span>
                  </div>
                </div>

                <div className="bg-gray-900 rounded-xl p-5 border border-gray-800">
                  <GaussianCurve
                    curvePoints={curvePoints}
                    betsOnNumberline={betsOnLine}
                    targetValue={targetValue}
                    unit={round.unit}
                    showTarget={true}
                  />
                </div>

                <div className="bg-gray-900 rounded-xl p-5 border border-gray-800">
                  <RankingsTable rankings={rankings} unit={round.unit} targetValue={targetValue} />
                </div>
              </>
            )}

            <div className="bg-gray-900 rounded-xl p-5 border border-gray-800">
              <h3 className="text-sm font-semibold text-gray-400 mb-3 uppercase tracking-wider">Debt Settlement</h3>
              <DebtChecklist debts={debts} myUsername={username} token={token} onUpdate={handleDebtUpdate} />
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
