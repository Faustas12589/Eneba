/**
 * Lithuanian labels for the backend's category ids (domain.Categories), in menu order.
 * The backend decides which products are in which category; this only names them.
 */
export const CATEGORIES = [
  { id: "games", label: "Žaidimai" },
  { id: "dlc", label: "DLC" },
  { id: "gaming-cards", label: "Žaidimų kortelės" },
  { id: "gift-cards", label: "eDovanų kortelės" },
  { id: "software", label: "Programinė įranga" },
];

/** Label for a category id; an unknown id is shown as is. */
export function categoryLabel(id: string): string {
  return CATEGORIES.find((c) => c.id === id)?.label ?? id;
}
