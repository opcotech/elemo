import Avatar from '@/components/Avatar';
import Page from '@/components/Page';
import { slugify } from '@/helpers/strings';

export const metadata = {
  title: 'Avatars | Styleguide | Elemo',
  description: 'Styleguide for Elemo'
};

type AvatarVariant = {
  size: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  src?: string;
  initials?: string;
};

const variants: { title: string; variants: AvatarVariant[] }[] = [
  {
    title: 'With images',
    variants: [
      { size: 'xs', src: 'https://picsum.photos/id/237/100/100' },
      { size: 'sm', src: 'https://picsum.photos/id/237/100/100' },
      { size: 'md', src: 'https://picsum.photos/id/237/100/100' },
      { size: 'lg', src: 'https://picsum.photos/id/237/100/100' },
      { size: 'xl', src: 'https://picsum.photos/id/237/100/100' }
    ]
  },
  {
    title: 'With initials',
    variants: [
      { size: 'xs', initials: 'TU' },
      { size: 'sm', initials: 'TU' },
      { size: 'md', initials: 'TU' },
      { size: 'lg', initials: 'TU' },
      { size: 'xl', initials: 'TU' }
    ]
  }
];

function AvatarContainer({ title, props }: { title: string; props: AvatarVariant }) {
  return (
    <div className="block">
      <div className={`w-24 h-24 mb-2 flex items-center justify-center`}>
        <Avatar {...props} />
      </div>
      <div className="text-sm text-center">{title}</div>
    </div>
  );
}

function AvatarRow({ title, variants }: { title: string; variants: AvatarVariant[] }) {
  return (
    <div id={`#${slugify(title)}`} className="mb-12">
      <h2 className="mb-6">
        <a href={`#${slugify(title)}`}>{title}</a>
      </h2>
      <div className="flex flex-row space-x-4 mb-2">
        {variants.map((props, i) => {
          return <AvatarContainer key={i} title={props.size} props={props} />;
        })}
      </div>
    </div>
  );
}

export default function AvatarsPage() {
  return (
    <Page title="Avatars">
      {variants.map((row, i) => (
        <AvatarRow key={i} title={row.title} variants={row.variants} />
      ))}
    </Page>
  );
}
