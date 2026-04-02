'use client'

import { useState, useEffect } from 'react'

type NumberEntry = {
  id: number
  value: number
  created_at: string
}

export default function Home() {
  const [numbers, setNumbers] = useState<NumberEntry[]>([])
  const [latest, setLatest] = useState<NumberEntry | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchNumbers = async () => {
    const res = await fetch('/api/numbers')
    const data = await res.json()
    setNumbers(data || [])
  }

  const generate = async () => {
    setLoading(true)
    const res = await fetch('/api/numbers', { method: 'POST' })
    const data = await res.json()
    setLatest(data)
    await fetchNumbers()
    setLoading(false)
  }

  useEffect(() => {
    fetchNumbers()
  }, [])

  return (
    <main className="min-h-screen bg-gray-950 text-white flex flex-col items-center justify-start p-12">
      <h1 className="text-4xl font-bold mb-8">Gaussian Pot</h1>

      <button
        onClick={generate}
        disabled={loading}
        className="bg-violet-600 hover:bg-violet-500 disabled:opacity-50 text-white font-semibold px-8 py-4 rounded-xl text-lg mb-8 transition-colors"
      >
        {loading ? 'Generating...' : 'Generate Number'}
      </button>

      {latest && (
        <div className="text-6xl font-mono font-bold text-violet-400 mb-12">
          {latest.value}
        </div>
      )}

      <div className="w-full max-w-md">
        <h2 className="text-lg font-semibold text-gray-400 mb-4">History</h2>
        {numbers.length === 0 ? (
          <p className="text-gray-600">No numbers generated yet.</p>
        ) : (
          <ul className="space-y-2">
            {numbers.map((n) => (
              <li
                key={n.id}
                className="flex justify-between bg-gray-900 px-4 py-3 rounded-lg"
              >
                <span className="font-mono text-violet-300">{n.value}</span>
                <span className="text-gray-500 text-sm">
                  {new Date(n.created_at).toLocaleTimeString()}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </main>
  )
}