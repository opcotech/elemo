import Page from '@/components/Page';
import {ContentSkeleton} from '@/components/Skeleton';

export const metadata = {
  title: 'Namespaces | Elemo'
};

export default function NamespacesPage() {
  return (
    <Page title="Namespaces">
      <ContentSkeleton />
    </Page>
  );
}
