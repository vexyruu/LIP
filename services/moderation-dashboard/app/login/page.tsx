import { Suspense } from "react";
import { LoginClient } from "./login-client";

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <main className="min-h-screen flex items-center justify-center">
          <p className="font-mono text-sm text-on-surface-variant uppercase tracking-widest">
            Loading…
          </p>
        </main>
      }
    >
      <LoginClient />
    </Suspense>
  );
}
