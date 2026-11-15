export function formatPrice(cents: number, currency = "EUR"): string {
  return new Intl.NumberFormat("lt-LT", { style: "currency", currency }).format(cents / 100);
}

const REGIONS: Record<string, string> = {
  GLOBAL: "Visas pasaulis",
  EUROPE: "Europa",
  "UNITED STATES": "JAV",
};

/** Lithuanian name of a seed region ("GLOBAL" -> "Visas pasaulis"); unknown ones as is. */
export function regionLabel(region: string): string {
  return REGIONS[region] ?? region;
}

/** "14:35" in local time, from an ISO timestamp. */
export function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString("lt-LT", { hour: "2-digit", minute: "2-digit" });
}

/** Lithuanian plural: 1 prekė, 2 prekės, 10 prekių, 21 prekė, 22 prekės. */
export function formatItemCount(n: number): string {
  const last = n % 10;
  const lastTwo = n % 100;
  if (last === 1 && lastTwo !== 11) return `${n} prekė`;
  if (last >= 2 && (lastTwo < 12 || lastTwo > 19)) return `${n} prekės`;
  return `${n} prekių`;
}
