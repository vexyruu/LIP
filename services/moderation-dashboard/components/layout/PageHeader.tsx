import type { ReactNode } from "react";
import type { BreadcrumbItem } from "@/lib/navigation";
import { Breadcrumbs } from "./Breadcrumbs";

type PageHeaderProps = {
  title: string;
  description?: string;
  breadcrumbs?: BreadcrumbItem[];
  actions?: ReactNode;
  meta?: ReactNode;
};

export function PageHeader({
  title,
  description,
  breadcrumbs,
  actions,
  meta,
}: PageHeaderProps) {
  return (
    <header className="page-header">
      {breadcrumbs && breadcrumbs.length > 0 && (
        <Breadcrumbs items={breadcrumbs} />
      )}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0 flex-1">
          <h1 className="page-title">{title}</h1>
          {description && <p className="page-description">{description}</p>}
          {meta && <div className="mt-3">{meta}</div>}
        </div>
        {actions && (
          <div className="flex shrink-0 flex-wrap items-center gap-3">
            {actions}
          </div>
        )}
      </div>
    </header>
  );
}
