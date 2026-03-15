import EmptyState from "./EmptyState";

export default function CardList({ title, items, renderItem }) {
  return (
    <div>
      <h3 className="mb-3 font-display text-lg font-semibold text-ink">
        {title}
      </h3>
      <div className="grid gap-3">
        {items.length === 0 ? (
          <EmptyState
            title={`No ${title.toLowerCase()} yet`}
            body="This section will populate as configuration grows."
            compact
          />
        ) : (
          items.map(renderItem)
        )}
      </div>
    </div>
  );
}
