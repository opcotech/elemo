import Button from '@/components/Button';
import Page from '@/components/Page';
import {slugify} from '@/helpers/strings';

export const metadata = {
  title: 'Buttons | Styleguide | Elemo',
  description: 'Styleguide for Elemo'
};

type ButtonVariant = {
  size: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  loading?: boolean;
  variant: 'primary' | 'secondary' | 'danger' | 'accent';
  disabled?: boolean;
};

const variants: { title: string; variants: ButtonVariant[] }[] = [
  {
    title: 'Sizes',
    variants: [
      {size: 'xs', variant: 'primary'},
      {size: 'sm', variant: 'primary'},
      {size: 'md', variant: 'primary'},
      {size: 'lg', variant: 'primary'},
      {size: 'xl', variant: 'primary'}
    ]
  },
  {
    title: 'Variants',
    variants: [
      {size: 'md', variant: 'primary'},
      {size: 'md', variant: 'secondary'},
      {size: 'md', variant: 'danger'},
      {size: 'md', variant: 'accent'}
    ]
  },
  {
    title: 'Loading',
    variants: [
      {size: 'md', loading: true, variant: 'primary'},
      {size: 'md', loading: true, variant: 'secondary'},
      {size: 'md', loading: true, variant: 'danger'},
      {size: 'md', loading: true, variant: 'accent'}
    ]
  },
  {
    title: 'Disabled',
    variants: [
      {size: 'md', disabled: true, variant: 'primary'},
      {size: 'md', disabled: true, variant: 'secondary'},
      {size: 'md', disabled: true, variant: 'danger'},
      {size: 'md', disabled: true, variant: 'accent'}
    ]
  }
];

function ButtonContainer({title, props}: { title: string; props: ButtonVariant }) {
  return (
    <div className="block">
      <div className={`w-24 h-24 mb-2 flex items-center justify-center`}>
        <Button {...props}>{props.variant}</Button>
      </div>
      <div className="text-sm text-center">{title}</div>
    </div>
  );
}

function ButtonRow({title, variants}: { title: string; variants: ButtonVariant[] }) {
  return (
    <div id={`#${slugify(title)}`} className="mb-12">
      <h2 className="mb-6">
        <a href={`#${slugify(title)}`}>{title}</a>
      </h2>
      <div className="flex flex-row space-x-4 mb-2">
        {variants.map((props, i) => {
          return <ButtonContainer key={i} title={props.size} props={props}/>;
        })}
      </div>
    </div>
  );
}

export default function ButtonsPage() {
  return (
    <Page title="Buttons">
      {variants.map((row, i) => (
        <ButtonRow key={i} title={row.title} variants={row.variants}/>
      ))}
    </Page>
  );
}
