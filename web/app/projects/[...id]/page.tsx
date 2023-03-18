import Page from '@/components/Page';
import { ContentSkeleton } from '@/components/Skeleton';
import Breadcrumb from '@/components/Breadcrumb';

export const metadata = {
  title: 'Projects | Elemo'
};

export default function ProjectsPage() {
  return (
    <>
      <Breadcrumb />
      <Page title="Projects">
        <ContentSkeleton />
      </Page>
    </>
  );
}
