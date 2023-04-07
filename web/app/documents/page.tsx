import Page from '@/components/Page';
import {ContentSkeleton} from '@/components/Skeleton';

export const metadata = {
  title: 'Documents | Elemo'
};

export default function DocumentsPage() {
  return (
    <Page title="Documents">
      <ContentSkeleton/>
    </Page>
  );
}
