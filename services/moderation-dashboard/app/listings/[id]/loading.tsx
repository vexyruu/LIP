export default function ListingDetailLoading() {
  return (
    <div className="page-shell animate-pulse" aria-busy="true" aria-live="polite">
      <div className="h-4 w-40 bg-surface-variant/60" />
      <div className="h-8 max-w-xl w-2/3 bg-surface-variant" />
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-12">
        <div className="xl:col-span-7 h-72 bg-surface-variant border border-surface-variant" />
        <div className="xl:col-span-5 space-y-4">
          <div className="h-24 bg-surface-variant border border-surface-variant" />
          <div className="h-24 bg-surface-variant border border-surface-variant" />
        </div>
      </div>
    </div>
  );
}
