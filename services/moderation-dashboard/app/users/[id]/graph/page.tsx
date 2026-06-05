import { GraphClient } from "./graph-client";

export default async function UserGraphPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <GraphClient userId={id} />;
}
