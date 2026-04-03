'use client';

type Debt = {
  id: number;
  payer_username: string;
  payee_username: string;
  amount: number;
  payer_confirmed: boolean;
  payee_confirmed: boolean;
  status: string;
};

type Props = {
  debts: Debt[];
  myUsername: string;
  token: string;
  onUpdate: (debt: Debt) => void;
};

export default function DebtChecklist({ debts, myUsername, token, onUpdate }: Props) {
  const pending = debts.filter(d => d.status === 'PENDING');

  if (pending.length === 0) {
    return (
      <div className="text-green-400 text-sm text-center py-4">
        All debts settled — you're clear!
      </div>
    );
  }

  async function confirm(debtID: number, endpoint: 'confirm-paid' | 'confirm-received') {
    const res = await fetch(`/api/debts/${debtID}/${endpoint}`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      const updated: Debt = await res.json();
      onUpdate(updated);
    }
  }

  return (
    <div className="space-y-3">
      {pending.map(d => {
        const isPayer = d.payer_username === myUsername;
        const isPayee = d.payee_username === myUsername;

        return (
          <div key={d.id} className="bg-gray-800 rounded-lg p-4 flex flex-col gap-2">
            <div className="flex justify-between text-sm">
              <span className="text-gray-300">
                <span className="text-red-400 font-medium">{d.payer_username}</span>
                {' → '}
                <span className="text-green-400 font-medium">{d.payee_username}</span>
              </span>
              <span className="font-mono text-white font-bold">{d.amount.toFixed(2)} cr</span>
            </div>

            <div className="flex gap-2 text-xs text-gray-500">
              <span className={d.payer_confirmed ? 'text-green-400' : ''}>
                {d.payer_confirmed ? '✓' : '○'} Payer confirmed
              </span>
              <span className={d.payee_confirmed ? 'text-green-400' : ''}>
                {d.payee_confirmed ? '✓' : '○'} Payee confirmed
              </span>
            </div>

            <div className="flex gap-2">
              {isPayer && !d.payer_confirmed && (
                <button
                  onClick={() => confirm(d.id, 'confirm-paid')}
                  className="flex-1 text-xs bg-red-700 hover:bg-red-600 text-white rounded py-1.5 transition-colors"
                >
                  I paid {d.payee_username}
                </button>
              )}
              {isPayee && !d.payee_confirmed && (
                <button
                  onClick={() => confirm(d.id, 'confirm-received')}
                  className="flex-1 text-xs bg-green-700 hover:bg-green-600 text-white rounded py-1.5 transition-colors"
                >
                  I received from {d.payer_username}
                </button>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
