export function formatPrice(cents: number, currency = "EUR"): string {
  return new Intl.NumberFormat("lt-LT", { style: "currency", currency }).format(cents / 100);
}

/** Lithuanian plural: 1 prekė, 2 prekės, 10 prekių, 21 prekė, 22 prekės. */
export function formatItemCount(n: number): string {
  const last = n % 10;
  const lastTwo = n % 100;
  if (last === 1 && lastTwo !== 11) return `${n} prekė`;
  if (last >= 2 && (lastTwo < 12 || lastTwo > 19)) return `${n} prekės`;
  return `${n} prekių`;
}
