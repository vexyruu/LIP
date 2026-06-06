import { getSession } from "@/lib/session";
import { canModerate } from "@/lib/rbac";
import { ListingDetailClient } from "./listing-detail-client";

export default async function ListingDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const session = await getSession();
  return (
    <ListingDetailClient
      listingId={id}
      moderatorId={session?.moderatorId ?? null}
      canModerate={session ? canModerate(session.role) : false}
    />
  );
}
