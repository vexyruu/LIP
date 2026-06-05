import Link from "next/link";

export default function UnauthorizedPage() {
  return (
    <div className="page-shell">
      <p className="font-mono text-sm uppercase tracking-widest text-alert-coral">
        Error 403
      </p>
      <h1 className="page-title">Unauthorized</h1>
      <p className="page-description">
        You do not have permission to view this resource. Sign in with a
        moderator account to continue.
      </p>
      <Link href="/login" className="btn-primary w-fit">
        Go to login
      </Link>
    </div>
  );
}
