import Page from '@/components/Page';
import Link from 'next/link';

export const metadata = {
  title: 'Styleguide | Elemo',
  description: 'Styleguide for Elemo'
};

export default function StyleguidePage() {
  return (
    <Page title="Styleguide">
      <ul className="list-disc">
        <li>
          <Link href="/styleguide/colors">Colors</Link>
        </li>
        <li>
          <Link href="/styleguide/avatars">Avatars</Link>
        </li>
        <li>
          <Link href="/styleguide/buttons">Buttons</Link>
        </li>
        <li>
          <Link href="/styleguide/spinners">Spinners</Link>
        </li>
      </ul>
    </Page>
  );
}
