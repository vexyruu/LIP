"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { PageHeader } from "@/components/layout/PageHeader";
import { SectionHeader } from "@/components/SectionHeader";
import { shortUUID } from "@/lib/format";
import type { GraphEdge, GraphNode, UserGraph } from "@/lib/types";

const NODE_TYPE_ORDER = ["User", "Device", "IPAddress", "PaymentMethod"];

function nodeLabel(node: GraphNode): string {
  if (node.type === "User") return shortUUID(node.id);
  return node.label;
}

function nodeRowClass(node: GraphNode, centerUserId: string): string {
  if (node.type === "User" && node.id === centerUserId) {
    return "border-l-2 border-alert-lime bg-surface-container-low";
  }
  if (node.banned) {
    return "border-l-2 border-alert-coral bg-alert-coral/5";
  }
  return "border-l-2 border-surface-variant bg-surface-container-lowest";
}

function formatNodeType(type: string): string {
  return type.replace(/([A-Z])/g, " $1").trim();
}

export function GraphClient({ userId }: { userId: string }) {
  const [graph, setGraph] = useState<UserGraph | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadGraph = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/api/users/${userId}/graph`);
      if (!res.ok) throw new Error(await res.text());
      setGraph((await res.json()) as UserGraph);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load graph");
      setGraph(null);
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    loadGraph();
  }, [loadGraph]);

  const nodesByType = useMemo(() => {
    if (!graph) return new Map<string, GraphNode[]>();
    const map = new Map<string, GraphNode[]>();
    for (const node of graph.nodes) {
      const list = map.get(node.type) ?? [];
      list.push(node);
      map.set(node.type, list);
    }
    return map;
  }, [graph]);

  if (loading && !graph) {
    return (
      <div className="page-shell">
        <p className="font-mono text-sm text-on-surface-variant">
          Loading connection graph…
        </p>
      </div>
    );
  }

  if (error || !graph) {
    return (
      <div className="page-shell">
        <PageHeader
          title="Connection graph"
          breadcrumbs={[
            { label: "Review queue", href: "/queue" },
            { label: "Seller profile", href: `/users/${userId}` },
            { label: "Graph" },
          ]}
        />
        <div className="status-banner status-banner-error" role="alert">
          {error ?? "No graph data"}
        </div>
        <button type="button" onClick={loadGraph} className="btn-secondary w-fit">
          Try again
        </button>
      </div>
    );
  }

  return (
    <div className="page-shell">
      <PageHeader
        title="Connection graph"
        description={`List view of Neo4j subgraph (D3 force graph out of scope) · ${shortUUID(userId)} · ${graph.nodes.length} nodes · ${graph.edges.length} edges`}
        breadcrumbs={[
          { label: "Review queue", href: "/queue" },
          { label: "Seller profile", href: `/users/${userId}` },
          { label: "Graph" },
        ]}
        actions={
          <Link href={`/users/${userId}`} className="btn-secondary">
            Back to profile
          </Link>
        }
      />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {NODE_TYPE_ORDER.map((type) => {
          const nodes = nodesByType.get(type) ?? [];
          if (nodes.length === 0) return null;
          return (
            <section key={type} className="panel overflow-hidden">
              <div className="panel-header">{formatNodeType(type)}</div>
              <ul className="divide-y divide-surface-variant">
                {nodes.map((node) => (
                  <li
                    key={`${node.type}-${node.id}`}
                    className={`p-3 ${nodeRowClass(node, userId)}`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <span className="break-all font-mono text-sm">
                        {nodeLabel(node)}
                      </span>
                      <div className="flex shrink-0 flex-col items-end gap-1">
                        {node.type === "User" && node.id !== userId && (
                          <Link
                            href={`/users/${node.id}`}
                            className="font-mono text-[10px] uppercase text-alert-lime underline-offset-2 hover:underline"
                          >
                            Profile
                          </Link>
                        )}
                        {node.banned && (
                          <span className="font-mono text-[10px] font-bold uppercase text-alert-coral">
                            Banned
                          </span>
                        )}
                        {node.risk_score != null && (
                          <span className="font-mono text-[10px] text-on-surface-variant">
                            Risk {node.risk_score.toFixed(2)}
                          </span>
                        )}
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </section>
          );
        })}
      </div>

      <section className="panel overflow-hidden">
        <div className="panel-header">Connections</div>
        {graph.edges.length === 0 ? (
          <p className="p-4 font-mono text-sm text-on-surface-variant">
            No edges in this subgraph.
          </p>
        ) : (
          <ul className="divide-y divide-surface-variant font-mono text-xs">
            {graph.edges.map((edge: GraphEdge, i) => (
              <li
                key={`${edge.source}-${edge.type}-${edge.target}-${i}`}
                className="flex flex-wrap items-center gap-2 p-3"
              >
                <span className="text-on-surface-variant">
                  {edge.type.replace(/_/g, " ")}
                </span>
                <span className="text-on-surface-variant">·</span>
                <span>{shortUUID(edge.source)}</span>
                <span className="text-alert-lime">→</span>
                <span>{shortUUID(edge.target)}</span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
