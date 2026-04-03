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

type Question = { id: number; text: string; target_value: number; unit: string };
type User = { id: number; username: string; role: string; credits: number };
type BetEntry = { username: string; guess: number; wager: number };
type CurvePoint = { x: number; y: number };
type RankEntry = { rank: number; username: string; guess: number; wager: number; payout: number; net: number };
type Debt = { id: number; payer_username: string; payee_username: string; amount: number; payer_confirmed: boolean; payee_confirmed: boolean; status: string };

export default function HostPage() {
  const router = useRouter();
  const [token, setToken] = useState('');
  const [username, setUsername] = useState('');

  const [gameState, setGameState] = useState<GameState>('IDLE');
  const [questions, setQuestions] = useState<Question[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [selectedQID, setSelectedQID] = useState<number | null>(null);
  const [selectedMode, setSelectedMode] = useState<'sniper' | 'social' | 'volatile'>('sniper');

  const [roundUnit, setRoundUnit] = useState('');
  const [roundQuestion, setRoundQuestion] = useState('');
  const [bets, setBets] = useState<BetEntry[]>([]);
  const [curvePoints, setCurvePoints] = useState<CurvePoint[]>([]);
  const [betsOnLine, setBetsOnLine] = useState<BetEntry[]>([]);
  const [rankings, setRankings] = useState<RankEntry[]>([]);
  const [targetValue, setTargetValue] = useState<number | null>(null);
  const [debts, setDebts] = useState<Debt[]>([]);

  const [replenishUser, setReplenishUser] = useState('');
  const [replenishAmt, setReplenishAmt] = useState('');
  const [actionMsg, setActionMsg] = useState('');

  useEffect(() => {
    const t = localStorage.getItem('gp_token');
    const u = localStorage.getItem('gp_username');
    const r = localStorage.getItem('gp_role');
    if (!t || r !== 'host') { router.push('/'); return; }
    setToken(t);
    setUsername(u || '');

    const headers = { Authorization: `Bearer ${t}` };
    fetch('/api/admin/questions', { headers }).then(r => r.json()).then(setQuestions);
    fetch('/api/admin/users', { headers }).then(r => r.json()).then(d => { if (Array.isArray(d)) setUsers(d); });
    fetch('/api/round', { headers }).then(r => r.json()).then(data => {
      if (data.state && data.state !== 'IDLE') {
        setGameState(data.state as GameState);
        setRoundQuestion(data.question_text || '');
        setRoundUnit(data.unit || '');
        if (data.target_value != null) setTargetValue(data.target_value);
      }
    });
  }, [router]);

  const refreshUsers = useCallback((t: string) => {
    fetch('/api/admin/users', { headers: { Authorization: `Bearer ${t}` } })
      .then(r => r.json()).then(d => { if (Array.isArray(d)) setUsers(d); });
  }, []);

  const handleEvent = useCallback((event: { type: string; payload: unknown }) => {
    const p = event.payload as Record<string, unknown>;
    switch (event.type) {
      case 'round_opened':
        setGameState('ROUND_OPEN');
        setRoundQuestion(p.question_text as string);
        setRoundUnit(p.unit as string);
        setBets([]);
        setCurvePoints([]);
        setBetsOnLine([]);
        setRankings([]);
        setTargetValue(null);
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
        setDebts((p.debts as Debt[]) || []);
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
        setDebts([]);
        break;

      case 'session_update': {
        const activeUsers = p.active_users as { username: string; role: string }[];
        setUsers(prev => prev.map(u => ({
          ...u,
          _online: activeUsers.some(a => a.username === u.username),
        })));
        break;
      }
    }
  }, []);

  useWebSocket(handleEvent);

  async function openRound() {
    if (!selectedQID) { setActionMsg('Select a question first'); return; }
    const res = await fetch('/api/admin/round/open', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ question_id: selectedQID, mode: selectedMode }),
    });
    const data = await res.json();
    if (!res.ok) setActionMsg(data.error || 'Failed to open round');
    else setActionMsg('');
  }

  async function callRound() {
    const res = await fetch('/api/admin/round/call', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    if (!res.ok) setActionMsg(data.error || 'Failed to call round');
    else setActionMsg('');
  }

  async function randomQuestion() {
    const res = await fetch('/api/admin/question/random', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    });
    const q: Question = await res.json();
    setSelectedQID(q.id);
  }

  async function replenish() {
    const res = await fetch('/api/admin/credits/replenish', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ username: replenishUser, amount: Number(replenishAmt) }),
    });
    if (res.ok) {
      setActionMsg(`Replenished ${replenishAmt} cr to ${replenishUser}`);
      setReplenishUser('');
      setReplenishAmt('');
      refreshUsers(token);
    } else {
      const d = await res.json();
      setActionMsg(d.error || 'Failed to replenish');
    }
  }

  async function logout() {
    await fetch('/api/auth/logout', { method: 'POST', headers: { Authorization: `Bearer ${token}` } });
    localStorage.clear();
    router.push('/');
  }

  const selectedQ = questions.find(q => q.id === selectedQID);
  const pendingDebts = debts.filter(d => d.status === 'PENDING');

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      <header className="sticky top-0 z-10 bg-gray-900 border-b border-gray-800 px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="font-bold text-violet-400">Gaussian Pot</span>
          <span className="text-xs text-yellow-400 bg-yellow-900/30 px-2 py-0.5 rounded">HOST</span>
          <span className="text-gray-500 text-sm">{username}</span>
        </div>
        <div className="flex items-center gap-3">
          <span className={`text-xs px-2 py-0.5 rounded font-medium ${
            gameState === 'IDLE' ? 'bg-gray-800 text-gray-400' :
            gameState === 'ROUND_OPEN' ? 'bg-green-900/40 text-green-400' :
            gameState === 'SETTLING' ? 'bg-yellow-900/40 text-yellow-400' :
            'bg-red-900/40 text-red-400'
          }`}>{gameState}</span>
          <button onClick={logout} className="text-xs text-gray-400 hover:text-white">Leave</button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto p-4 grid grid-cols-1 lg:grid-cols-3 gap-4">

        {/* Left panel: controls */}
        <div className="lg:col-span-1 space-y-4">

          {/* Session panel */}
          <div className="bg-gray-900 rounded-xl p-4 border border-gray-800">
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Players</h3>
            {users.filter(u => u.role === 'player').length === 0 ? (
              <p className="text-gray-600 text-sm">No players yet</p>
            ) : (
              <div className="space-y-2">
                {users.filter(u => u.role === 'player').map(u => (
                  <div key={u.id} className="flex justify-between items-center text-sm">
                    <span className="text-white">{u.username}</span>
                    <span className="font-mono text-green-400 text-xs">{u.credits.toFixed(2)} cr</span>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Replenish credits */}
          <div className="bg-gray-900 rounded-xl p-4 border border-gray-800">
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Replenish Credits</h3>
            <div className="space-y-2">
              <select
                value={replenishUser}
                onChange={e => setReplenishUser(e.target.value)}
                className="w-full bg-gray-800 text-white rounded px-3 py-1.5 text-sm border border-gray-700"
              >
                <option value="">Select player</option>
                {users.filter(u => u.role === 'player').map(u => (
                  <option key={u.id} value={u.username}>{u.username}</option>
                ))}
              </select>
              <input
                type="number"
                value={replenishAmt}
                onChange={e => setReplenishAmt(e.target.value)}
                placeholder="Amount"
                className="w-full bg-gray-800 text-white rounded px-3 py-1.5 text-sm border border-gray-700"
              />
              <button
                onClick={replenish}
                disabled={!replenishUser || !replenishAmt}
                className="w-full bg-yellow-700 hover:bg-yellow-600 disabled:bg-gray-700 disabled:text-gray-500 text-white text-xs font-semibold rounded py-1.5 transition-colors"
              >
                Add Credits
              </button>
            </div>
          </div>

          {/* Round control */}
          <div className="bg-gray-900 rounded-xl p-4 border border-gray-800 space-y-3">
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Round Control</h3>

            {gameState === 'IDLE' && (
              <>
                <div>
                  <label className="text-xs text-gray-500 mb-1 block">Mode</label>
                  <div className="flex gap-1">
                    {(['sniper', 'social', 'volatile'] as const).map(m => (
                      <button
                        key={m}
                        onClick={() => setSelectedMode(m)}
                        className={`flex-1 text-xs py-1.5 rounded capitalize transition-colors ${
                          selectedMode === m ? 'bg-violet-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
                        }`}
                      >
                        {m}
                      </button>
                    ))}
                  </div>
                </div>

                <div>
                  <div className="flex justify-between items-center mb-1">
                    <label className="text-xs text-gray-500">Question</label>
                    <button onClick={randomQuestion} className="text-xs text-violet-400 hover:text-violet-300">Random</button>
                  </div>
                  <select
                    value={selectedQID ?? ''}
                    onChange={e => setSelectedQID(Number(e.target.value))}
                    className="w-full bg-gray-800 text-white rounded px-3 py-1.5 text-sm border border-gray-700"
                    size={1}
                  >
                    <option value="">Select question…</option>
                    {questions.map(q => (
                      <option key={q.id} value={q.id}>{q.id}. {q.text.substring(0, 50)}</option>
                    ))}
                  </select>
                </div>

                <button
                  onClick={openRound}
                  disabled={!selectedQID}
                  className="w-full bg-green-700 hover:bg-green-600 disabled:bg-gray-700 disabled:text-gray-500 text-white font-semibold rounded py-2 text-sm transition-colors"
                >
                  Open Round
                </button>
              </>
            )}

            {gameState === 'ROUND_OPEN' && (
              <button
                onClick={callRound}
                className="w-full bg-red-700 hover:bg-red-600 text-white font-semibold rounded py-2 text-sm transition-colors"
              >
                Call Round
              </button>
            )}

            {gameState === 'SETTLING' && (
              <p className="text-xs text-yellow-400 text-center">
                Waiting for {pendingDebts.length} debt{pendingDebts.length !== 1 ? 's' : ''} to settle…
              </p>
            )}

            {actionMsg && <p className="text-xs text-red-400">{actionMsg}</p>}
          </div>
        </div>

        {/* Right panel: live monitor */}
        <div className="lg:col-span-2 space-y-4">

          {gameState === 'ROUND_OPEN' && (
            <>
              <div className="bg-gray-900 rounded-xl p-4 border border-gray-800">
                <p className="text-xs text-gray-500 mb-1 uppercase tracking-wider">Active Question</p>
                <p className="text-white font-medium">{roundQuestion}</p>
              </div>
              <div className="bg-gray-900 rounded-xl p-4 border border-gray-800 space-y-3">
                <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Live Distribution (Host View — Target Visible)</h3>
                {selectedQ && (
                  <GaussianCurve
                    curvePoints={curvePoints}
                    betsOnNumberline={betsOnLine}
                    targetValue={selectedQ.target_value}
                    unit={roundUnit}
                    showTarget={true}
                  />
                )}
                <BubbleChart bets={betsOnLine} targetValue={selectedQ?.target_value} unit={roundUnit} />
              </div>
              <div className="bg-gray-900 rounded-xl p-4 border border-gray-800">
                <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Bet Ticker ({bets.length} bets)</h3>
                <BetTicker bets={bets} unit={roundUnit} />
              </div>
            </>
          )}

          {(gameState === 'SETTLING' || gameState === 'ROUND_CALLED') && targetValue != null && (
            <>
              <div className="bg-gray-900 rounded-xl p-4 border border-gray-800">
                <p className="text-xs text-gray-500 mb-1">{roundQuestion}</p>
                <div className="flex items-baseline gap-2">
                  <span className="text-3xl font-bold font-mono text-red-400">{targetValue}</span>
                  <span className="text-gray-400">{roundUnit}</span>
                </div>
              </div>
              <div className="bg-gray-900 rounded-xl p-4 border border-gray-800">
                <GaussianCurve curvePoints={curvePoints} betsOnNumberline={betsOnLine} targetValue={targetValue} unit={roundUnit} showTarget={true} />
              </div>
              <div className="bg-gray-900 rounded-xl p-4 border border-gray-800">
                <RankingsTable rankings={rankings} unit={roundUnit} targetValue={targetValue} />
              </div>
              <div className="bg-gray-900 rounded-xl p-4 border border-gray-800">
                <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">All Debts</h3>
                <DebtChecklist debts={debts} myUsername={username} token={token} onUpdate={d => setDebts(prev => prev.map(x => x.id === d.id ? d : x))} />
              </div>
            </>
          )}

          {gameState === 'IDLE' && (
            <div className="bg-gray-900 rounded-xl p-8 border border-gray-800 text-center text-gray-500">
              <p>Select a question and open a round to begin.</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
