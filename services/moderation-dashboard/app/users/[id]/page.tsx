import { getSession } from "@/lib/session";
import { canModerate } from "@/lib/rbac";
import { UserClient } from "./user-client";

export default async function UserPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const session = await getSession();
  return (
    <UserClient
      userId={id}
      canModerate={session ? canModerate(session.role) : false}
    />
  );
}
