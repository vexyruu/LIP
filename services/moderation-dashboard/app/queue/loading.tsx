export default function QueueLoading() {
  return (
    <div className="page-shell animate-pulse" aria-busy="true" aria-live="polite">
      <div className="h-8 w-48 bg-surface-variant" />
      <div className="h-4 w-64 bg-surface-variant/60" />
      <div className="panel p-4">
        <div className="flex gap-2">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-10 w-20 bg-surface-variant" />
          ))}
        </div>
      </div>
      <div className="panel space-y-3 p-4">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="flex gap-4 border-b border-surface-variant/50 py-4">
            <div className="h-10 w-10 bg-surface-variant" />
            <div className="flex-1 space-y-2">
              <div className="h-4 w-1/3 bg-surface-variant" />
              <div className="h-3 w-1/4 bg-surface-variant/60" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
