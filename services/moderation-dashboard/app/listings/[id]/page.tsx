import { ListingDetailClient } from "./listing-detail-client";

export default async function ListingDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <ListingDetailClient listingId={id} />;
}
