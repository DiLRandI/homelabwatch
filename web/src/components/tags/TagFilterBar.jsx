import Badge from "../ui/Badge";
import Button from "../ui/Button";

export default function TagFilterBar({
  onSelect,
  selectedTag = "",
  tags = [],
}) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <Button
        onClick={() => onSelect("")}
        size="sm"
        variant={selectedTag ? "ghost" : "secondary"}
      >
        All tags
      </Button>
      {tags.map((tag) => (
        <button
          className={`inline-flex h-9 items-center gap-2 rounded-lg border px-3 text-sm font-medium transition ${
            selectedTag === tag.name
              ? "border-accent bg-accent text-white"
              : "border-line bg-panel-strong text-ink-soft hover:border-line-strong hover:text-ink"
          }`}
          key={tag.id}
          onClick={() => onSelect(tag.name)}
          type="button"
        >
          <span>{tag.name}</span>
          <Badge>{tag.bookmarkCount}</Badge>
        </button>
      ))}
    </div>
  );
}
