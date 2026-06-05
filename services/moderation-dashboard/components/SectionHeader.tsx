export function SectionHeader({ title }: { title: string }) {
  return (
    <div className="section-header">
      <span className="text-on-surface-variant text-sm tracking-widest uppercase font-bold">
        {title}
      </span>
    </div>
  );
}
