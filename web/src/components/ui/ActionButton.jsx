export default function ActionButton({ children, onClick }) {
  return (
    <button
      className="rounded-full border border-accent/40 bg-base/70 px-4 py-3 text-sm font-semibold text-accent transition hover:bg-accent hover:text-white"
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );
}
