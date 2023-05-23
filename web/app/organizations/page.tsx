import Page from '@/components/Page';
import { ListSkeleton } from '@/components/Skeleton';
import OrganizationList from './OrganizationList';

export const metadata = {
  title: 'Organizations | Elemo'
};

export default function SiteSettingsPage() {
  return (
    <Page title="Organizations">
      <OrganizationList />
    </Page>
  );
}
