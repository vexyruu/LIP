import Link from "next/link";

export default function NotFound() {
  return (
    <div className="page-shell">
      <p className="font-mono text-sm uppercase tracking-widest text-alert-coral">
        Error 404
      </p>
      <h1 className="page-title">Page not found</h1>
      <p className="page-description">
        This route does not exist in the moderation dashboard. Check the URL or
        return to the review queue.
      </p>
      <Link href="/queue" className="btn-primary w-fit">
        Back to queue
      </Link>
    </div>
  );
}
