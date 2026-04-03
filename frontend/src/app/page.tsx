'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [tooMany, setTooMany] = useState(false);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError('');
    setTooMany(false);

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (res.status === 429) {
        setTooMany(true);
        return;
      }

      const data = await res.json();
      if (!res.ok) {
        setError(data.error || 'Login failed');
        return;
      }

      localStorage.setItem('gp_token', data.token);
      localStorage.setItem('gp_username', data.username);
      localStorage.setItem('gp_role', data.role);
      localStorage.setItem('gp_credits', String(data.credits));

      if (data.role === 'host') {
        router.push('/host');
      } else {
        router.push('/player');
      }
    } catch {
      setError('Connection failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-10">
          <h1 className="text-4xl font-bold text-white tracking-tight">Gaussian Pot</h1>
          <p className="mt-2 text-gray-400 text-sm">Real-time sports betting on a bell curve</p>
        </div>

        {tooMany && (
          <div className="mb-6 rounded-lg bg-yellow-900/40 border border-yellow-700 p-4 text-yellow-200 text-sm text-center">
            Too many players at the moment — try again later.
          </div>
        )}

        <form onSubmit={handleLogin} className="bg-gray-900 rounded-2xl p-8 space-y-5 shadow-xl border border-gray-800">
          <div>
            <label className="block text-xs text-gray-400 mb-1 uppercase tracking-wider">Username</label>
            <input
              type="text"
              value={username}
              onChange={e => setUsername(e.target.value)}
              placeholder="your name"
              required
              className="w-full bg-gray-800 text-white rounded-lg px-4 py-3 text-sm outline-none border border-gray-700 focus:border-violet-500 transition-colors"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-400 mb-1 uppercase tracking-wider">
              Password <span className="text-gray-600 normal-case">(host only)</span>
            </label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              placeholder="leave blank for players"
              className="w-full bg-gray-800 text-white rounded-lg px-4 py-3 text-sm outline-none border border-gray-700 focus:border-violet-500 transition-colors"
            />
          </div>

          {error && (
            <p className="text-red-400 text-sm text-center">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading || !username}
            className="w-full bg-violet-600 hover:bg-violet-500 disabled:bg-gray-700 disabled:text-gray-500 text-white font-semibold rounded-lg py-3 transition-colors"
          >
            {loading ? 'Joining…' : 'Join Game'}
          </button>
        </form>

        <p className="mt-6 text-center text-xs text-gray-600">
          Max 8 players + 1 host. Credits reset each session.
        </p>
      </div>
    </div>
  );
}
