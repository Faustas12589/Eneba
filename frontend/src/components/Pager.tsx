interface Props {
  page: number;
  pageSize: number;
  total: number;
  onPage: (page: number) => void;
}

/** Previous / next page buttons for a paged API response; hidden when it all fits on one page. */
export function Pager({ page, pageSize, total, onPage }: Props) {
  if (total <= pageSize) return null;
  const pages = Math.ceil(total / pageSize);
  return (
    <div className="pager">
      <button className="btn secondary" disabled={page <= 1} onClick={() => onPage(page - 1)}>Ankstesnis</button>
      <span className="muted">Puslapis {page} / {pages}</span>
      <button className="btn secondary" disabled={page >= pages} onClick={() => onPage(page + 1)}>Kitas</button>
    </div>
  );
}
