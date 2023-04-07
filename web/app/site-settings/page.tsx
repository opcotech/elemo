import Page from '@/components/Page';
import {ContentSkeleton} from '@/components/Skeleton';

export const metadata = {
  title: 'Site settings | Elemo'
};

export default function SiteSettingsPage() {
  return (
    <Page title="Site settings">
      <ContentSkeleton/>
    </Page>
  );
}
