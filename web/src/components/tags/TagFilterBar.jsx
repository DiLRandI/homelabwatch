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
          className={`inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-sm font-medium transition ${
            selectedTag === tag.name
              ? "border-accent bg-accent text-white"
              : "border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:text-slate-950"
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
