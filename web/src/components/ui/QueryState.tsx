export function QueryState({ pending, error }: { pending: boolean; error: Error | null }) {
  return <p role={error ? 'alert' : 'status'} className="rounded-lg border border-slate-200 bg-white p-6 text-slate-700">
    {pending ? 'Loading…' : error?.message}
  </p>;
}
