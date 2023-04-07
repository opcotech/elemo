import Page from '@/components/Page';
import Breadcrumb from '@/components/Breadcrumb';
import Link from '@/components/Link';

export const metadata = {
  title: 'Projects | Elemo'
};

export default function ProjectsPage() {
  return (
    <>
      <Breadcrumb />
      <Page title="Projects">
        <ul>
          <li>
            <Link href="/projects/webapp">projects/webapp</Link>
          </li>
          <li>
            <Link href="/projects/webapp/releases">projects/webapp/releases</Link>
          </li>
          <li>
            <Link href="/projects/webapp/releases/v1.0">projects/webapp/releases/v1.0</Link>
          </li>
          <li>
            <Link href="/projects/webapp/releases/v1.0/issues">projects/webapp/releases/v1.0/issues</Link>
          </li>
          <li>
            <Link href="/projects/webapp/releases/v1.0/issues/WEB-1">projects/webapp/releases/v1.0/issues/WEB-1</Link>
          </li>
        </ul>
      </Page>
    </>
  );
}
