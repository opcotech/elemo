'use client';

import useStore from "@/store";
import { use, useEffect } from "react";

export default function OrganizationList() {
  const organizations = useStore((state) => state.organizations);
  const fetchedOrganizations = useStore((state) => state.fetchedOrganizations);
  const fetchingOrganizations = useStore((state) => state.fetchingOrganizations);
  const fetchOrganizations = useStore((state) => state.fetchOrganizations);

  useEffect(() => {
    if (!fetchingOrganizations && !fetchedOrganizations && organizations.length === 0) {
      Promise.resolve(fetchOrganizations());
    }
  }, [organizations, fetchedOrganizations, fetchingOrganizations, fetchOrganizations]);

  return (
    <ul>
      {organizations.map((organization) => (
        <li key={organization.id}>{organization.name} ({organization.members.length})</li>
      ))}
    </ul>
  )
}
