"use client";

import { useEffect, useRef } from "react";

type BanModalProps = {
  open: boolean;
  loading: boolean;
  onClose: () => void;
  onConfirm: (reason: string) => void;
};

export function BanModal({
  open,
  loading,
  onClose,
  onConfirm,
}: BanModalProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (open) {
      textareaRef.current?.focus();
    }
  }, [open]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-[200] flex items-center justify-center bg-black/80 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="ban-modal-title"
    >
      <div className="w-full max-w-lg border-2 border-white bg-background p-8">
        <h2
          id="ban-modal-title"
          className="text-2xl font-bold uppercase tracking-tight"
        >
          Ban user
        </h2>
        <p className="text-on-surface-variant mt-2 text-sm">
          This marks the seller as banned in Postgres and the fraud graph.
          Reason is required.
        </p>
        <textarea
          ref={textareaRef}
          id="ban-reason"
          rows={4}
          placeholder="e.g. Direct link to banned fraud ring"
          className="mt-6 w-full border-2 border-surface-variant bg-surface-container-low p-4 font-mono text-sm text-on-surface placeholder:text-on-surface-variant/50 focus:border-alert-lime focus:outline-none resize-none"
          disabled={loading}
        />
        <div className="mt-8 flex gap-4">
          <button
            type="button"
            onClick={onClose}
            disabled={loading}
            className="flex-1 h-12 border-2 border-surface-variant font-bold text-xs uppercase tracking-widest hover:border-white transition-colors disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={loading}
            onClick={() => {
              const reason = textareaRef.current?.value.trim() ?? "";
              if (!reason) {
                textareaRef.current?.focus();
                return;
              }
              onConfirm(reason);
            }}
            className="flex-1 h-12 bg-alert-coral text-background font-black text-xs uppercase tracking-widest brutalist-border-bold hover:brightness-110 disabled:opacity-50"
          >
            {loading ? "Banning…" : "Confirm ban"}
          </button>
        </div>
      </div>
    </div>
  );
}
