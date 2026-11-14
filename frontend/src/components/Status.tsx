import type { AsyncState } from "../lib/useAsync";

/** Shared loading / error wrapper so every block handles its own failure. */
export function Status<T>({ state, children }: { state: AsyncState<T>; children: (data: T) => React.ReactNode }) {
  if (state.loading && !state.data) return <p className="muted">Kraunama…</p>;
  if (state.error) return <p className="error">Nepavyko įkelti: {state.error}</p>;
  if (!state.data) return null;
  return <>{children(state.data)}</>;
}
