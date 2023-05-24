import type { StateCreator } from 'zustand';
import client, {
  ContentType,
  getErrorMessage,
  OrganizationStatus,
  V1OrganizationsGetParams,
  V1OrganizationUpdateData
} from '@/lib/api';
import type { Organization } from '@/lib/api';
import type { MessageSliceState } from './messageSlice';

export type CreateOrganizationInput = {
  name: string;
  email: string;
  logo?: string;
  website?: string;
};

export type UpdateOrganizationInput = {
  name?: string;
  email?: string;
  logo?: string;
  website?: string;
  status?: OrganizationStatus;
};

export interface OrganizationState {
  organization: Organization | null;
  fetchingOrganization: boolean;
  fetchedOrganization: boolean;
  fetchOrganization: (id: string) => Promise<void>;
  createOrganization: (organization: CreateOrganizationInput) => Promise<void>;
  updateOrganization: (id: string, organization: UpdateOrganizationInput) => Promise<void>;
  deleteOrganization: (id: string) => Promise<void>;

  organizations: Organization[];
  fetchingOrganizations: boolean;
  fetchedOrganizations: boolean;
  fetchOrganizations: (params?: V1OrganizationsGetParams) => Promise<void>;

  organizationMembers: string[];
  fetchingOrganizationMembers: boolean;
  fetchedOrganizationMembers: boolean;
  fetchOrganizationMembers: (id: string) => Promise<void>;
  addOrganizationMember: (id: string, user_id: string) => Promise<void>;
  removeOrganizationMember: (id: string, user_id: string) => Promise<void>;
}

export const createOrganizationSlice: StateCreator<OrganizationState & Partial<MessageSliceState>> = (set, get) => ({
  organization: null,
  fetchingOrganization: false,
  fetchedOrganization: false,
  fetchOrganization: async (id: string) => {
    try {
      set({ fetchingOrganization: true });
      const res = await client.v1.v1OrganizationGet(id);
      const organization: Organization = await res.json();
      set({ organization });
      set({ fetchingOrganization: false, fetchedOrganization: true });
    } catch (e) {
      set({ fetchingOrganization: false });
      get().addMessage?.({
        type: 'error',
        title: 'Failed to fetch organization',
        message: getErrorMessage(e)
      });
    }
  },
  createOrganization: async (organization: CreateOrganizationInput) => {
    try {
      const res = await client.v1.v1OrganizationsCreate(organization, { type: ContentType.Json });
      const data: { id: string } = await res.json();

      // NOTE: This is an ugly hack, the API should return the created organization
      // instead of just the ID.
      set((state) => ({
        organizations: [
          ...state.organizations,
          {
            id: data.id,
            name: organization.name,
            email: organization.email,
            logo: organization.logo || '',
            website: organization.website || '',
            status: OrganizationStatus.Active,
            members: [],
            teams: [],
            namespaces: [],
            created_at: new Date().toISOString(),
            updated_at: null
          }
        ]
      }));

      get().addMessage?.({
        type: 'success',
        title: 'Organization created',
        message: `Organization "${organization.name}" created successfully`
      });
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to create organization',
        message: getErrorMessage(e)
      });
    }
  },
  updateOrganization: async (id: string, organization: UpdateOrganizationInput) => {
    try {
      const res = await client.v1.v1OrganizationUpdate(id, organization, { type: ContentType.Json });
      const updated: V1OrganizationUpdateData = await res.json();
      set((state) => ({
        organizations: state.organizations.map((organization) => (organization.id === id ? updated : organization))
      }));
      get().addMessage?.({
        type: 'success',
        title: 'Organization updated',
        message: `Organization "${id}" updated successfully`
      });
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to update organization',
        message: getErrorMessage(e)
      });
    }
  },
  deleteOrganization: async (id: string) => {
    try {
      const res = await client.v1.v1OrganizationDelete({ id });
      await res.json();
      set((state) => ({ organizations: state.organizations.filter((org) => org.id !== id) }));
      get().addMessage?.({
        type: 'success',
        title: 'Organization deleted',
        message: `Organization "${id}" deleted successfully`
      });
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to delete organization',
        message: getErrorMessage(e)
      });
    }
  },

  organizations: [],
  fetchingOrganizations: false,
  fetchedOrganizations: false,
  fetchOrganizations: async (params: V1OrganizationsGetParams = {}) => {
    try {
      set({ fetchingOrganizations: true });
      const res = await client.v1.v1OrganizationsGet(params);
      const organizations: Organization[] = await res.json();
      set({ organizations });
      set({ fetchingOrganizations: false, fetchedOrganizations: true });
    } catch (e) {
      set({ fetchingOrganizations: false });
      return get().addMessage?.({
        type: 'error',
        title: 'Failed to fetch organizations',
        message: getErrorMessage(e)
      });
    }
  },

  organizationMembers: [],
  fetchingOrganizationMembers: false,
  fetchedOrganizationMembers: false,
  fetchOrganizationMembers: async (id: string) => {
    try {
      set({ fetchingOrganizationMembers: true });
      const res = await client.v1.v1OrganizationMembersGet(id);
      const organizationMembers: string[] = await res.json();
      set({ organizationMembers });
      set({ fetchingOrganizationMembers: false, fetchedOrganizationMembers: true });
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to fetch organization members',
        message: getErrorMessage(e)
      });
    }
  },
  addOrganizationMember: async (id: string, user_id: string) => {
    try {
      const res = await client.v1.v1OrganizationMembersAdd(id, { user_id }, { type: ContentType.Json });
      await res.json();
      get().addMessage?.({
        type: 'success',
        title: 'Organization member added',
        message: `Organization member "${user_id}" added successfully`
      });

      // Refetch organization members, this could be optimized by just adding the new member
      // to the existing list if the addition would return the new member.
      await get().fetchOrganizationMembers(id);
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to add organization member',
        message: getErrorMessage(e)
      });
    }
  },
  removeOrganizationMember: async (id: string, user_id: string) => {
    try {
      const res = await client.v1.v1OrganizationMembersRemove(id, user_id, { type: ContentType.Json });
      await res.json();
      get().addMessage?.({
        type: 'success',
        title: 'Organization member removed',
        message: `Organization member "${user_id}" removed successfully`
      });
      set((state) => ({ organizationMembers: state.organizationMembers.filter((member) => member !== user_id) }));
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to remove organization member',
        message: getErrorMessage(e)
      });
    }
  }
});
