import Page from '@/components/Page';
import {ContentSkeleton} from '@/components/Skeleton';

export const metadata = {
  title: 'Settings | Elemo'
};

export default function SettingsPage() {
  return (
    <Page title="Settings">
      <ContentSkeleton/>
    </Page>
  );
}
